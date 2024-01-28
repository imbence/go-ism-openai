package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log"
	"strings"
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
	defer func(state bool, config Config) {
		err := connectDB(state, config)
		if err != nil {

		}
	}(false, config)

	type Person struct {
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
		Email     string
	}

	personStructs := []Person{
		{FirstName: "Ardie", LastName: "Savea", Email: "aaaaaaa"},
		{FirstName: "Sonny Bill", LastName: "Williams", Email: "bbbbbbbb"},
		{FirstName: "Ngani", LastName: "Laumape", Email: "cccccc"},
	}
	var res sql.Result
	var query string = `INSERT INTO person (first_name, last_name, email) VALUES (:first_name, :last_name, :email) ON CONFLICT (first_name) DO UPDATE SET first_name = excluded.first_name, last_name  = excluded.last_name, email = excluded.email`
	res, err = db.NamedExec(query, personStructs)

	println(res.RowsAffected())
	println(res.LastInsertId())

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
	err := db.Select(&rankReport, query)
	if err != nil {
		log.Fatal("Error querying database: " + err.Error())
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

func listenDbUpdates() error {
	// Listen for updates to the ai_request table
	// Establish a PostgreSQL connection
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", config.DB.Host, config.DB.Port, config.DB.User, config.DB.Pass, config.DB.DBname)
	db, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		return err
	}
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal("Error closing the database:", err)
		}
	}(db)

	// Enable LISTEN/NOTIFY feature for real-time notifications
	listenChannel := make(chan string)
	listener := pq.NewListener(psqlInfo, 10*time.Second, time.Minute, func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Println("Listener error:", err)
		}
	})
	defer func(listener *pq.Listener) {
		err := listener.Close()
		if err != nil {
			log.Fatal("Error closing the listener:", err)
		}
	}(listener)

	err = listener.Listen("tasks_update")
	if err != nil {
		log.Fatal("Error setting up LISTEN/NOTIFY:", err)
	}

	log.Println("Start monitoring PostgreSQL...")
	go func() {
		for {
			select {
			case n := <-listener.Notify:
				// Notification received, execute your code here
				fmt.Printf("Received notification: %+v\n", n)
				listenChannel <- n.Channel
			case <-time.After(5 * time.Second):
				// Re-establish the listen connection after 5 seconds
				err := listener.Ping()
				if err != nil {
					log.Println("Error pinging listener:", err)
					return
				}
			}
		}
	}()

	// Handle incoming signals to gracefully exit the program
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-signalChan:
			fmt.Printf("Received signal %s. Exiting...\n", signalChan)
			os.Exit(0)
		case notification := <-listenChannel:
			// Execute your code here for each notification
			fmt.Printf("Received notification: %s\n", notification)
		}
	}
}
