package main

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

type Report struct {
	Date    string `db:"date"`
	Part    string `db:"part"`
	Content string `db:"content"`
}

type AiIndustryRanks struct {
	Date        string `json:"date"`
	Industry    string `json:"industry"`
	Part        string `json:"part"`
	Rank        int    `json:"rank"`
	AiRequestID string `json:"ai_request_id"`
	//Comment     string `json:"comment"`
}

type PrimaryKey struct {
	ColumnName string `json:"column_name"`
}

type Dates struct {
	Date string `json:"date"`
}
