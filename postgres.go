package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
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

func toDBb(columnnames []string, schema string, table string, rowvalues []string, primarykey string) error {
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

func toDB(schema string, table string, data interface{}) error {

	var err error
	var columnName []string
	var primaryKey []PrimaryKey
	var primaryKeyColumn []string
	//var cellValue []string
	//var rowValue []string

	primaryKeyQuery := fmt.Sprintf("select json_agg(x) from (SELECT column_name FROM information_schema.key_column_usage WHERE table_name = '%s' and table_schema = '%s') x", table, schema)
	err = dbQuery(primaryKeyQuery, &primaryKey)
	if err != nil {
		log.Println("Error querying database: " + err.Error())
		return err
	}

	//build primary key string
	for i := range primaryKey {
		primaryKeyColumn = append(primaryKeyColumn, primaryKey[i].ColumnName)
	}
	primaryKeyString := strings.Join(primaryKeyColumn, ", ")
	onConflictKeyString := "excluded." + strings.Join(primaryKeyColumn, ", excluded.")

	//get column names from struct
	t := reflect.TypeOf(data)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		columnName = append(columnName, field.Name)
	}
	columnNameString := strings.Join(columnName, ", ")

	// todo: or this
	t = reflect.TypeOf(data)
	v := reflect.ValueOf(data)

	// Loop through the fields of the struct
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		columnName = append(columnName, field.Name)
		value := v.Field(i).Interface()

		fmt.Printf("%s: %v\n", field.Name, value)
	}

	//put together the query
	sqlStatement := fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES %s ON CONFLICT (%s) DO UPDATE SET %s", schema, table, columnNameString, "", primaryKeyString, onConflictKeyString)

	//send query to database
	_, err = db.Exec(sqlStatement)
	if err != nil {
		log.Println("Send query to database failed: " + err.Error() + " Table: " + table)
	}

	return err
}

func dbQuery(query string, queryResult interface{}) error {

	rows, err := db.Query(query)
	if err != nil {
		log.Println("Error querying database: " + err.Error())
		return err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			return //error already logged
		}
	}(rows)

	var cellValue string

	if rows.Next() {
		err := rows.Scan(&cellValue)
		if err != nil {
			return err
		}
		err = json.Unmarshal([]byte(cellValue), &queryResult)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("No rows found")
	}
	return nil
}
