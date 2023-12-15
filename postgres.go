package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var (
	db *sql.DB
)

func connectDB(state bool, config Config) error {
	var err error
	if state {
		psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", config.DB.Host, config.DB.Port, config.DB.User, config.DB.Pass, config.DB.DBname)
		db, err = sql.Open("postgres", psqlInfo)
		if err != nil {
			log.Println(err)
			return err
		}
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)
		db.SetConnMaxIdleTime(30 * time.Minute)
		return nil
	} else {
		err = db.Close()
		return err
	}
}

func toDB(columnnames []string, schema string, table string, rowvalues []string, primarykey string) error {
	//https://golangdocs.com/golang-postgresql-example
	var err error
	var onconflict string
	var allrows = strings.Join(rowvalues, ", ")
	var allcolumnnames = strings.ToLower(strings.Join(columnnames, ", "))

	//build on conflict string
	for i := range columnnames {
		onconflict += columnnames[i] + " = excluded." + columnnames[i] + ", "
	}
	onconflict = strings.ToLower(strings.TrimSuffix(onconflict, ", "))

	//put together the query
	sqlStatement := fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES %s ON CONFLICT (%s) DO UPDATE SET %s", schema, table, allcolumnnames, allrows, primarykey, onconflict)

	//send query to database
	_, err = db.Exec(sqlStatement)
	if err != nil {
		log.Println("Send query to database failed: " + err.Error() + " Table: " + table)
	}

	return err
}

func getManReports() []ManReport {

	var manReports []ManReport
	query := "SELECT json_agg(man_reports) FROM ism.man_reports"
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(rows)

	var cellValue string
	if rows.Next() {
		err := rows.Scan(&cellValue)
		if err != nil {
			log.Println(err)
		}
		err = json.Unmarshal([]byte(cellValue), &manReports)
		if err != nil {
			log.Println("Unmarshal json from DB: " + err.Error())
		}
	} else {
		fmt.Println("No rows found")
	}

	return manReports
}

func dbQuery(query string, queryResult interface{}) interface{} {

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(rows)

	var cellValue string
	if rows.Next() {
		err := rows.Scan(&cellValue)
		if err != nil {
			log.Println(err)
		}
		err = json.Unmarshal([]byte(cellValue), &queryResult)
		if err != nil {
			log.Println("Unmarshal json from DB: " + err.Error())
		}
	} else {
		fmt.Println("No rows found")
	}

	return queryResult
}
