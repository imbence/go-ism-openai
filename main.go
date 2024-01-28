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

	// todo: start the magic from Grafana
	// todo: implement db listener

	//Run the AI on the reports
	var rankReportPart []Report
	getReportData("ranking3", []string{"2022-02-01"}, &rankReportPart)
	runAiOnReports(rankReportPart)

	// Run the AI on the comments
	var commentReportPart []Report
	getReportData("respondents1", []string{"2024-01-02", "2023-12-04", "2023-11-01", "2023-10-02", "2023-09-01", "2023-08-01", "2023-07-03", "2023-06-01", "2023-05-01", "2023-04-03", "2023-03-01", "2023-02-01", "2023-01-04", "2022-12-01", "2022-11-01", "2022-10-03", "2022-09-01", "2022-08-01", "2022-07-01", "2022-06-01", "2022-05-02", "2022-04-01", "2022-03-01", "2022-02-01", "2022-01-04"}, &commentReportPart)
	runAiOnComments(commentReportPart)

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

func runAiOnReports(rankReportPart []Report) {
	var err error

	// Do the AI magic
	var aiResponse []AiResponse
	var aiIndustryRanks []AiIndustryRanks
	for i := range rankReportPart {
		fmt.Println("AI magic for: ", rankReportPart[i].Date, rankReportPart[i].Part)
		aiResponse = append(aiResponse, AiMagic(rankReportPart[i].Content))

		if err := json.Unmarshal([]byte(aiResponse[i].Choices[0].Message.Content), &aiIndustryRanks); err != nil {
			log.Println("Error unmarshalling json: " + err.Error())
		}
		for x := range aiIndustryRanks {
			aiIndustryRanks[x].Date = rankReportPart[i].Date
			aiIndustryRanks[x].Part = rankReportPart[i].Part
			aiIndustryRanks[x].AiRequestID = rankReportPart[i].AiRequestID
		}
		// Send to database
		err = toDB("ism", rankReportPart[i].TargetTable, aiIndustryRanks)
		if err != nil {
			log.Println("Error sending to database: " + err.Error())
		}
		// Clear the slice
		aiIndustryRanks = aiIndustryRanks[:0]
	}
}

func runAiOnComments(commentReportPart []Report) {
	var err error
	// Do the AI magic
	var aiResponse []AiResponse
	var aiComments []AiComments
	for i := range commentReportPart {
		fmt.Println("AI magic for: ", commentReportPart[i].Date, commentReportPart[i].Part)
		aiResponse = append(aiResponse, AiMagic(commentReportPart[i].Content))

		if err := json.Unmarshal([]byte(aiResponse[i].Choices[0].Message.Content), &aiComments); err != nil {
			log.Println("Error unmarshalling json: " + err.Error())
		}
		var aiCommentsNullsRemoved []AiComments
		for x := range aiComments {
			aiComments[x].Date = commentReportPart[i].Date
			aiComments[x].AiRequestID = commentReportPart[i].AiRequestID
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
}
