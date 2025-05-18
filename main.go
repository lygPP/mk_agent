package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"lygPP/mk_agent/ai/aiclient/deepseek"
	"lygPP/mk_agent/ai/httputil"
	"lygPP/mk_agent/ai/mcp"
	"lygPP/mk_agent/model"
	"net/http"
	"os"
)

func main() {
	ctx := context.Background()
	cfg := httputil.Config{
		URL:        "https://ark.cn-beijing.volces.com/api/v1/chat/completions",
		APIKey:     "55eeebc0-ef98-49a5-8404-e65aa40903f4",
		HTTPClient: &http.Client{},
	}
	aiClient := deepseek.NewDeepSeek(cfg)
	body := model.Request{
		Model:    "ep-20250220181854-c8s82",
		Messages: []model.Message{{Role: "system", Content: "You are a helpful assistant."}},
	}

	// 获取mcp tool信息
	tools, err := mcp.NewMkMcpClient(ctx).GenDeeepseekTools(ctx)
	if err != nil {
		fmt.Printf("get mcp tools err: %+v", err)
	} else {
		body.Tools = tools
	}

	for {
		fmt.Printf("输入你的问题：")
		reader := bufio.NewReader(os.Stdin)
		userInput, _ := reader.ReadString('\n')
		if userInput == "q" {
			break
		}
		body.Messages = append(body.Messages, model.Message{
			Role:    "user",
			Content: userInput,
		})

		for i := 0; i < 3; i++ {
			needCallTool := false
			toolIndexToInfoMap := make(map[int32]map[string]map[string]string)

			// bodyStr, _ := json.Marshal(body)
			// fmt.Printf("body: %s \n", bodyStr)

			// // 同步调用
			// res, err := aiClient.Chat(context.Background(), body)
			// if err != nil {
			// 	fmt.Printf("%+v", err)
			// 	panic(err)
			// }
			// if len(res.ToolCalls) > 0 {
			// 	needCallTool = true
			// 	toolIndexToInfoMap[0] = map[string]map[string]string{
			// 		"function": {
			// 			"name":      res.ToolCalls[0]["function"].(map[string]any)["name"].(string),
			// 			"arguments": res.ToolCalls[0]["function"].(map[string]any)["arguments"].(string),
			// 		},
			// 	}
			// }
			// fmt.Println(res.Content)

			// 流式调用
			allResContent := ""
			resChannel, _ := aiClient.Stream(context.Background(), body)
			k := 0
			y := 0
			for res := range resChannel {
				if len(res.ToolCalls) > 0 {
					needCallTool = true
					for _, tmpTool := range res.ToolCalls {
						if toolIndexToInfoMap[int32(tmpTool["index"].(float64))] == nil {
							toolIndexToInfoMap[int32(tmpTool["index"].(float64))] = map[string]map[string]string{
								"function": {
									"name":      "",
									"arguments": "",
									// "id":        "",
								},
							}
						}
						if tmpTool["function"] != nil {
							tmpFunction := tmpTool["function"].(map[string]any)
							if tmpFunction["arguments"] != nil && tmpFunction["arguments"].(string) != "" {
								toolIndexToInfoMap[int32(tmpTool["index"].(float64))]["function"]["arguments"] = toolIndexToInfoMap[int32(tmpTool["index"].(float64))]["function"]["arguments"] + tmpFunction["arguments"].(string)
							}
							if tmpFunction["name"] != nil && tmpFunction["name"].(string) != "" {
								toolIndexToInfoMap[int32(tmpTool["index"].(float64))]["function"]["name"] = toolIndexToInfoMap[int32(tmpTool["index"].(float64))]["function"]["name"] + tmpFunction["name"].(string)
							}
						}
						// if tmpTool["id"] != nil {
						// 	tmpFunctionId := tmpTool["id"].(string)
						// 	if tmpFunctionId != "" {
						// 		toolIndexToInfoMap[int32(tmpTool["index"].(float64))]["function"]["id"] = toolIndexToInfoMap[int32(tmpTool["index"].(float64))]["function"]["id"] + tmpFunctionId
						// 	}
						// }
					}
				}
				if res.ReasoningContent != "" {
					if k == 0 {
						fmt.Printf("Reasoning：")
					}
					k++
					fmt.Printf("%s", res.ReasoningContent)
					allResContent = allResContent + res.ReasoningContent
					continue
				}
				if res.Content != "" {
					if y == 0 {
						fmt.Printf("最终结果：")
					}
					y++
					fmt.Printf("%s", res.Content)
				}
			}

			// 开始function call
			if needCallTool {
				functionInfo := toolIndexToInfoMap[0]["function"]
				toolName := functionInfo["name"]
				toolArguments := make(map[string]any)
				fmt.Printf("调用工具：%s \n", toolName)
				_ = json.Unmarshal([]byte(functionInfo["arguments"]), &toolArguments)
				toolResult, err := mcp.NewMkMcpClient(ctx).CallTool(ctx, toolName, toolArguments)
				if err != nil {
					fmt.Printf("call mcp tool err: %+v", err)
					break
				} else {
					body.Messages = append(body.Messages, model.Message{
						Role:    "assistant",
						Content: fmt.Sprintf("调用外部工具 %s 获取结果", functionInfo["name"]),
					})
					body.Messages = append(body.Messages, model.Message{
						Role:       "tool",
						Content:    toolResult,
						ToolCallId: functionInfo["name"],
					})
					body.Tools = nil
				}
			} else {
				break
			}
		}
	}
}
