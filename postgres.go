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

func toDB(schema string, table string, data interface{}) error {

	var err error
	var columnName []string
	var columnNameString string
	var primaryKey []PrimaryKey
	var primaryKeyColumn []string
	var primaryKeyString string
	var onConflictKeyString string
	var cellValue []string
	var rowValue []string
	var rowValuesString string

	//get column names
	sliceValue := reflect.ValueOf(data)
	if sliceValue.Kind() != reflect.Slice {
		return fmt.Errorf("data is not a slice")
	}

	// get column names and values
	for i := 0; i < sliceValue.Len(); i++ {
		item := sliceValue.Index(i)
		if item.Kind() == reflect.Struct {
			dataType := item.Type()
			cellValue = cellValue[:0]
			for j := 0; j < dataType.NumField(); j++ {
				field := dataType.Field(j)
				if i == 0 {
					columnName = append(columnName, field.Tag.Get("json"))
				}
				cellValue = append(cellValue, fmt.Sprintf("'%v'", item.FieldByName(field.Name)))
			}
			rowValue = append(rowValue, "("+strings.Join(cellValue, ", ")+")")
		}
	}
	columnNameString = strings.ToLower(strings.Join(columnName, ", "))
	rowValuesString = strings.Join(rowValue, ", ")

	//get primary key
	primaryKeyQuery := fmt.Sprintf("SELECT column_name FROM information_schema.key_column_usage WHERE table_name = '%s' and table_schema = '%s'", table, schema)
	err = dbQuery(primaryKeyQuery, &primaryKey)
	if err != nil {
		log.Println("Error finding target table key columns in database: " + err.Error())
		return err
	}
	//build primary key string
	for i := range primaryKey {
		primaryKeyColumn = append(primaryKeyColumn, primaryKey[i].ColumnName)
	}
	primaryKeyString = strings.Join(primaryKeyColumn, ", ")

	//build on conflict string
	for i := range primaryKeyColumn {
		primaryKeyColumn[i] = primaryKeyColumn[i] + " = excluded." + primaryKeyColumn[i]
	}
	onConflictKeyString = strings.Join(primaryKeyColumn, ", ")

	//put together the query
	sqlStatement := fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES %s ON CONFLICT (%s) DO UPDATE SET %s", schema, table, columnNameString, rowValuesString, primaryKeyString, onConflictKeyString)

	//send query to database
	_, err = db.Exec(sqlStatement)
	if err != nil {
		log.Println("Send query to database failed: " + err.Error() + " Table: " + table)
		log.Println(sqlStatement)
	}

	return err
}

func dbQuery(query string, queryResult interface{}) error {

	rows, err := db.Query(fmt.Sprintf("select json_agg(x) from(%s) x", query))
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
