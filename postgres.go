package main

import (
	"context"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	"log"
	"reflect"
	"strings"
)

func toDB(schema string, table string, data interface{}) error {

	type PrimaryKey struct {
		ColumnName string `db:"column_name"`
	}

	var err error
	var columnName []string
	var columnNameString string
	var primaryKey []PrimaryKey
	var primaryKeyColumn []string
	var primaryKeyString string
	var onConflictColumn []string
	var onConflictKeyString string
	var cellValue []string
	var rowValue []string
	var rowValuesString string

	//get column names
	sliceValue := reflect.ValueOf(data)
	if sliceValue.Kind() == reflect.Struct {
		dataType := reflect.TypeOf(data)
		dataValue := reflect.ValueOf(data)
		for i := 0; i < dataType.NumField(); i++ {
			field := dataType.Field(i)
			columnName = append(columnName, field.Tag.Get("db"))
			cellValue := fmt.Sprintf("$$%v$$", dataValue.Field(i))
			rowValue = append(rowValue, cellValue)
		}
		rowValuesString = "(" + strings.Join(rowValue, ", ") + ")"
	} else if sliceValue.Kind() == reflect.Slice {
		// get column names and values
		for i := 0; i < sliceValue.Len(); i++ {
			item := sliceValue.Index(i)
			if item.Kind() == reflect.Struct {
				dataType := item.Type()
				cellValue = cellValue[:0]
				for j := 0; j < dataType.NumField(); j++ {
					field := dataType.Field(j)
					if i == 0 {
						columnName = append(columnName, field.Tag.Get("db"))
					}
					cellValue = append(cellValue, fmt.Sprintf("$$%v$$", item.FieldByName(field.Name)))
				}
				rowValue = append(rowValue, "("+strings.Join(cellValue, ", ")+")")
			}
		}
		rowValuesString = strings.Join(rowValue, ", ")
	}
	columnNameString = strings.ToLower(strings.Join(columnName, ", "))

	//get primary key
	primaryKeyQuery := fmt.Sprintf("SELECT column_name FROM information_schema.key_column_usage WHERE table_name = '%s' and table_schema = '%s'", table, schema)
	err = pgxscan.Select(context.Background(), db, &primaryKey, primaryKeyQuery)
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
	for i := range columnName {
		onConflictColumn = append(onConflictColumn, columnName[i]+" = excluded."+columnName[i])
	}
	onConflictKeyString = strings.Join(onConflictColumn, ", ")

	//put together the query
	sqlStatement := fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES %s ON CONFLICT (%s) DO UPDATE SET %s", schema, table, columnNameString, rowValuesString, primaryKeyString, onConflictKeyString)

	//send query to database
	_, err = db.Exec(context.Background(), sqlStatement)
	if err != nil {
		log.Println(sqlStatement)
		log.Println("Send query to database failed: " + err.Error() + " Table: " + table)
	}

	return err
}
