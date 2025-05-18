package model

import "context"

type Model interface {
	Chat(ctx context.Context, req Request) (*Response, error)
	Stream(ctx context.Context, req Request) (<-chan Response, error)
}

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
	Tools    []Tool    `json:"tools"`
}

type Message struct {
	Role       string `json:"role"` // system、user、assistant、tool
	Content    string `json:"content"`
	ToolCallId string `json:"tool_call_id"`
}

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

/*
*

	"function": {
		"name": "get_weather",
		"description": "Get weather of an location, the user shoud supply a location first",
		"parameters": {
			"type": "object",
			"properties": {
				"location": {
					"type": "string",
					"description": "The city and state, e.g. San Francisco, CA",
				}
			},
			"required": ["location"]
		},
	}
*/
type Function struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type Response struct {
	Content          string           `json:"content"`
	ReasoningContent string           `json:"reasoning_content"` // 思考回答
	ToolCalls        []map[string]any `json:"tool_calls"`
}
