package main

import (
	"log"
)

func main() {
	config, _ = LoadConfiguration("config.json")

	//connect to the database
	err := connectDB(true, config)
	if err != nil {
		log.Println("Error connecting to database: " + err.Error())
	}

	var manReports []ManReport
	manReports = getManReports()

	var requestText AiRequest
	requestText = dbQuery("SELECT json_agg(ai_request) FROM ism.ai_request where id = $$ranking2$$", requestText).(AiRequest)

	//run the AI
	aiResult := AiMagic(requestText[0].Content + manReports[0].BacklogOfOrdersRank)

	println(aiResult.Choices[0].Message.Content)

	//close the database
	err = connectDB(false, config)
	if err != nil {
		log.Println("Error closing database: " + err.Error())
	}
}
