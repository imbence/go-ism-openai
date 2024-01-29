package main

type AiResponse struct {
	ID        string `db:"id"`
	Object    string `db:"object"`
	CreatedAt int    `db:"created_at"`
	Choices   []struct {
		Index   int `db:"index"`
		Message struct {
			Role    string `db:"role"`
			Content string `db:"content"`
		} `db:"message"`
		FinishReason string `db:"finish_reason"`
	} `db:"choices"`
	Usage struct {
		PromptTokens     int `db:"prompt_tokens"`
		CompletionTokens int `db:"completion_tokens"`
		TotalTokens      int `db:"total_tokens"`
	} `db:"usage"`
}

type Report struct {
	Date        string `db:"date"`
	Part        string `db:"part"`
	Content     string `db:"content"`
	TargetTable string `db:"target_table"`
	AiRequestID string `db:"ai_request_id"`
}

type AiIndustryRanks struct {
	Date        string `db:"date"`
	Industry    string `db:"industry"`
	Part        string `db:"part"`
	Rank        int    `db:"rank"`
	AiRequestID string `db:"ai_request_id"`
	TaskID      string `db:"task_id"`
}

type AiComments struct {
	Date        string `db:"date"`
	Industry    string `db:"industry"`
	Comment     string `db:"comment"`
	AiRequestID string `db:"ai_request_id"`
	TaskID      string `db:"task_id"`
}

type AiTasks struct {
	TaskID         string `db:"task_id"`
	AiRequestID    string `db:"ai_request_id"`
	TargetTable    string `db:"target_table"`
	AiRequestDates string `db:"ai_request_dates"`
}

type AiTasksUpdate struct {
	TaskID       string `db:"task_id"`
	AiStatus     string `db:"ai_status"`
	AiStartDate  string `db:"ai_start_date"`
	AiFinishDate string `db:"ai_finish_date"`
	AiMeta       string `db:"ai_meta"`
}
