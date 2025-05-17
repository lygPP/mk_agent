package main

import (
	"context"
	"fmt"
	"lygPP/mk_agent/ai/aiclient/deepseek"
	"lygPP/mk_agent/ai/httputil"
	"lygPP/mk_agent/model"
	"net/http"
)

func main() {
	cfg := httputil.Config{
		URL:        "https://ark.cn-beijing.volces.com/api/v1/chat/completions",
		APIKey:     "55eeebc0-ef98-49a5-8404-e65aa40903f4",
		HTTPClient: &http.Client{},
	}
	aiClient := deepseek.NewDeepSeek(cfg)
	body := model.Request{
		Model:    "ep-20250220181854-c8s82",
		Messages: []model.Message{{Role: "system", Content: "You are a helpful assistant."}, {Role: "user", Content: "你是谁"}},
	}
	// 同步调用
	chat, err := aiClient.Chat(context.Background(), body)
	if err != nil {
		fmt.Printf("%+v", err)
		panic(err)
	}
	fmt.Println(chat.Content)
	// 流式调用
	resChannel, _ := aiClient.Stream(context.Background(), body)
	for res := range resChannel {
		if res.ReasoningContent != "" {
			fmt.Println(res.ReasoningContent)
			continue
		}
		fmt.Println(res.Content)
	}
}
