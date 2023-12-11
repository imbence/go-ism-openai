package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ayush6624/go-chatgpt"
	"log"
	"reflect"
)

type AiResponse struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	CreatedAt int    `json:"created_at"`
	Choices   []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func main() {
	var config Config
	config, _ = LoadConfiguration("config.json")

	//connect to the database
	err := connectDB(true, config)
	if err != nil {
		log.Println("Error connecting to database: " + err.Error())
	}

	// this function is where we connect to the database and get the data we need into ManReports struct

	query := "SELECT date, \"into\", overall_rank, respondents, commodities, manufacturing_pmi, new_orders, new_orders_rank, production, production_rank, employment, employment_rank, supplier_deliveries, supplier_deliveries_rank, inventories, inventories_rank, customers_inventories, customers_inventories_rank, prices, prices_rank, backlog_of_orders, backlog_of_orders_rank, new_export_orders, new_export_orders_rank, imports, imports_rank, buying_policy FROM ism.man_reports"

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		log.Fatal(err)
	}

	structType := reflect.TypeOf(ManReports{})
	results := make([]ManReports, 0)

	for rows.Next() {
		data := reflect.New(structType).Elem()

		values := make([]interface{}, len(columns))
		for i := range columns {
			field := data.FieldByName(columns[i])
			if !field.IsValid() {
				log.Fatalf("Field %s not found in struct", columns[i])
			}
			values[i] = field.Addr().Interface()
		}

		if err := rows.Scan(values...); err != nil {
			log.Fatal(err)
		}

		results = append(results, data.Interface().(ManReports))
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Println(results)

	println("haha")
	println(rows)
	//rows to struct

	//run the AI
	AiMagic(config)

	//close the database
	err = connectDB(false, config)
	if err != nil {
		log.Println("Error closing database: " + err.Error())
	}
}

func AiMagic(config Config) {
	c, err := chatgpt.NewClient(config.ApiKeys.OpenaiApikey)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	res, err := c.Send(ctx, &chatgpt.ChatCompletionRequest{
		Model: chatgpt.GPT4,
		Messages: []chatgpt.ChatMessage{
			{
				Role:    chatgpt.ChatGPTModelRoleSystem,
				Content: "Hey, Explain GoLang to me in 2 sentences.",
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	choicesJSON, err := json.Marshal(res)
	if err != nil {
		log.Fatal(err)
	}

	var aiResponse AiResponse
	if err := json.Unmarshal(choicesJSON, &aiResponse); err != nil {
		log.Fatal(err)
	}

	println(aiResponse.Choices[0].Message.Content)
}
