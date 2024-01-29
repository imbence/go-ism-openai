package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"strings"
	_ "strings"
	"time"
	_ "time"
)

var (
	// db *pgx.Conn
	db *pgxpool.Pool
)

func main() {
	//Load the configuration
	var err error
	config, err = LoadConfiguration("config.json")
	if err != nil {
		log.Println("Error loading configuration: " + err.Error())
	}
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", config.DB.User, config.DB.Pass, config.DB.Host, config.DB.Port, config.DB.DBname)

	// Connect to the database
	db, err = pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		log.Fatal("Error connecting to database: " + err.Error())
	}
	defer db.Close()

	// Listen to the channel
	listenToChannel("run_tasks")
}

func listenToChannel(channelName string) {
	conn, err := db.Acquire(context.Background())
	if err != nil {
		log.Println("Error acquiring connection from pool: ", err)
		return
	}
	defer conn.Release()

	_, err = conn.Exec(context.Background(), "LISTEN "+channelName)
	if err != nil {
		log.Println("Error listening to channel: ", err)
		return
	}

	log.Println("Listening to channel", channelName)
	for {
		notification, err := conn.Conn().WaitForNotification(context.Background())
		if err != nil {
			log.Println("Error waiting for notification: ", err)
			return
		}
		log.Println("Received notification from channel", notification.Channel+":", notification.Payload)
		var aiTask []AiTasks
		err = pgxscan.Select(context.Background(), db, &aiTask, `select task_id, ai_request_id, target_table, array_to_string(ai_request_dates, ', ') as ai_request_dates from ism.ai_tasks where ai_status = 'CREATED'`)
		if err != nil {
			log.Fatal("Error querying database, AI Tasks Table: " + err.Error())
		} else if len(aiTask) > 0 {
			executeTask(aiTask)
		} else {
			log.Println("No AI Tasks to execute")
		}
	}
}

func executeTask(aiTask []AiTasks) {
	var err error
	// Loop through the AI Tasks
	for i := range aiTask {
		log.Printf("AI Task Loaded with ID: %s RequestID: %s Target Table: %s For Dates: %s", aiTask[i].TaskID, aiTask[i].AiRequestID, aiTask[i].TargetTable, aiTask[i].AiRequestDates)

		// Get the report data
		var reportPart []Report
		var aiTasksUpdate AiTasksUpdate
		aiTasksUpdate.AiMeta = "{}"
		getReportData(aiTask[i].AiRequestID, strings.Split(aiTask[i].AiRequestDates, ","), &reportPart)

		// Run the AI
		aiTasksUpdate.AiStartDate = time.Now().Format("2006-01-02 15:04:05")
		switch aiTask[i].TargetTable {
		case "ai_industry_ranks":
			err = runAiOnReports(reportPart, aiTask[i].TaskID)
		case "ai_comments":
			err = runAiOnComments(reportPart, aiTask[i].TaskID)
		}

		// Update the AI Tasks table
		aiTasksUpdate.TaskID = aiTask[i].TaskID
		if err != nil {
			log.Println("Error running AI: " + err.Error())
			aiTasksUpdate.AiStatus = "FAILED"
			aiTasksUpdate.AiMeta = fmt.Sprintf(`{"error": "%s"}`, err.Error())
		} else {
			aiTasksUpdate.AiStatus = "FINISHED"
		}
		aiTasksUpdate.AiFinishDate = time.Now().Format("2006-01-02 15:04:05")

		// Send to database
		err = toDB("ism", "ai_tasks", aiTasksUpdate)
		if err != nil {
			log.Println("Error sending to database: " + err.Error())
		} else {
			log.Println("AI Task Finished with ID: ", aiTasksUpdate.TaskID)
		}
		aiTasksUpdate = AiTasksUpdate{}
	}
}

func getReportData(aiRequestID string, reportDate []string, reportPart interface{}) {
	var err error
	var reportDates = "'" + strings.Join(reportDate, "', '") + "'"
	var sqlStatement = fmt.Sprintf(
		`select date::text, part, ar.content || ' ' || r.content as content, target_table, id as ai_request_id
				FROM ism.reports r 
				LEFT JOIN ism.ai_request ar ON r.part = ANY (ar.target_part) 
				WHERE r.part = ANY (ar.target_part) and id = $$%s$$ and date in (%s)`,
		aiRequestID, reportDates)

	// Query the database
	err = pgxscan.Select(context.Background(), db, reportPart, sqlStatement)
	if err != nil {
		log.Println(sqlStatement)
		log.Fatal("Error querying database: " + err.Error())
	}
}

func runAiOnReports(rankReportPart []Report, taskID string) error {
	var err error
	// Do the AI magic
	var aiResponse []AiResponse
	var aiIndustryRanks []AiIndustryRanks
	for i := range rankReportPart {
		log.Println("AI magic for: ", rankReportPart[i].Date, rankReportPart[i].Part)
		aiResponse = append(aiResponse, AiMagic(rankReportPart[i].Content))

		if err := json.Unmarshal([]byte(aiResponse[i].Choices[0].Message.Content), &aiIndustryRanks); err != nil {
			log.Println("Error unmarshalling json: " + err.Error())
		}
		for x := range aiIndustryRanks {
			aiIndustryRanks[x].Date = rankReportPart[i].Date
			aiIndustryRanks[x].Part = rankReportPart[i].Part
			aiIndustryRanks[x].AiRequestID = rankReportPart[i].AiRequestID
			aiIndustryRanks[x].TaskID = taskID
		}
		// Send to database
		err = toDB("ism", rankReportPart[i].TargetTable, aiIndustryRanks)
		if err != nil {
			log.Println("Error sending to database: " + err.Error())
		}
		// Clear the slice
		aiIndustryRanks = aiIndustryRanks[:0]
	}
	return err
}

func runAiOnComments(commentReportPart []Report, taskID string) error {
	var err error
	// Do the AI magic
	var aiResponse []AiResponse
	var aiComments []AiComments
	for i := range commentReportPart {
		log.Println("AI magic for: ", commentReportPart[i].Date, commentReportPart[i].Part)
		aiResponse = append(aiResponse, AiMagic(commentReportPart[i].Content))

		if err := json.Unmarshal([]byte(aiResponse[i].Choices[0].Message.Content), &aiComments); err != nil {
			log.Println("Error unmarshalling json: " + err.Error())
		}
		var aiCommentsNullsRemoved []AiComments
		for x := range aiComments {
			aiComments[x].Date = commentReportPart[i].Date
			aiComments[x].AiRequestID = commentReportPart[i].AiRequestID
			aiComments[x].TaskID = taskID
			if aiComments[x].Comment != "" {
				aiCommentsNullsRemoved = append(aiCommentsNullsRemoved, aiComments[x])
			}
		}
		// Send to database
		err = toDB("ism", commentReportPart[i].TargetTable, aiCommentsNullsRemoved)
		if err != nil {
			log.Println("Error sending to database: " + err.Error())
		}
		// Clear the slice
		aiComments = aiComments[:0]
	}
	return err
}
