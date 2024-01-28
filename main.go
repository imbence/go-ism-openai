package main

import (
	"encoding/json"
	"fmt"
	//"github.com/jmoiron/sqlx"
	//"github.com/lib/pq"
	"log"
	//"os"
	//"os/signal"
	"strings"
	//"syscall"
	//"time"
)

func main() {
	//Load the configuration
	var err error
	config, err = LoadConfiguration("config.json")
	if err != nil {
		log.Println("Error loading configuration: " + err.Error())
	}

	// Connect to the database
	err = connectDB(true, config)
	if err != nil {
		log.Println("Error connecting to database: " + err.Error())
	}

	// todo: get dates from database
	var dates []string = []string{"2022-02-01"}
	var reportDates = "'" + strings.Join(dates, "', '") + "'"
	var aiRequestID string = "ranking3"

	// Run the AI on the reports
	runAiOnReports(reportDates, aiRequestID)

	// Close the database connection
	err = connectDB(false, config)
	if err != nil {
		log.Println("Error closing database: " + err.Error())
	}
}

func runAiOnReports(reportDates string, aiRequestID string) {
	// Get the reports
	var rankReport []Report
	var query string = fmt.Sprintf("select date, part, ar.content || ' ' || r.content as content FROM ism.reports r LEFT JOIN ism.ai_request ar ON r.part = ANY (ar.target_part) WHERE r.part = ANY (ar.target_part) and id = $$%s$$ and date in (%s)", aiRequestID, reportDates)
	err := dbQuery(query, &rankReport)
	if err != nil {
		log.Println("Error querying database: " + err.Error())
	}

	// Do the AI magic
	var aiResponse []AiResponse
	var aiIndustryRanks []AiIndustryRanks
	for i := range rankReport {
		fmt.Println("AI magic for: ", rankReport[i].Date, rankReport[i].Part)
		aiResponse = append(aiResponse, AiMagic(rankReport[i].Content))

		if err := json.Unmarshal([]byte(aiResponse[i].Choices[0].Message.Content), &aiIndustryRanks); err != nil {
			log.Println("Error unmarshalling json: " + err.Error())
		}
		for x := range aiIndustryRanks {
			aiIndustryRanks[x].Date = rankReport[i].Date
			aiIndustryRanks[x].Part = rankReport[i].Part
			aiIndustryRanks[x].AiRequestID = aiRequestID
		}

		err = toDB("ism", "ai_industry_ranks", aiIndustryRanks)
		if err != nil {
			log.Println("Error sending to database: " + err.Error())
		}

		// Clear the slice
		aiIndustryRanks = aiIndustryRanks[:0]
	}
}
