package main

import (
	"database/sql"
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

func fromDB(columname string, schema string, table string, otherargs string) []string {
	var cellvalues []string

	//put together the query
	sqlStatement := "SELECT " + columname + " FROM " + schema + "." + table + " " + otherargs

	//send query to database
	rows, err := db.Query(sqlStatement)
	if err != nil {
		log.Println("DB Query failed: " + err.Error())
	}

	//close the rows
	defer func(rows *sql.Rows) {
		err = rows.Close()
		if err != nil {
			log.Println(err)
		}
	}(rows)

	//get the values
	for rows.Next() {
		var cellvalue string
		err = rows.Scan(&cellvalue)
		if err != nil {
			log.Println(err)
		}
		cellvalues = append(cellvalues, cellvalue)
	}
	return cellvalues
}
