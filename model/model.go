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
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Response struct {
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content"` // 思考回答
}
