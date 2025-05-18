package httputil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"lygPP/mk_agent/model"
	"net/http"
)

func ParseEvent(line []byte) *model.Response {
	// 处理事件流格式
	if !bytes.HasPrefix(line, []byte("data: ")) {
		return nil
	}
	payload := bytes.TrimPrefix(line, []byte("data: "))
	if string(payload) == "[DONE]" {
		return nil
	}

	// 解析响应结构
	var chunk struct {
		Choices []struct {
			Delta struct {
				Content          string           `json:"content"`
				ReasoningContent string           `json:"reasoning_content"`
				ToolCalls        []map[string]any `json:"tool_calls"`
			}
			FinishReason string `json:"finish_reason"`
		}
	}

	if err := json.Unmarshal(payload, &chunk); err != nil {
		return nil
	}

	if len(chunk.Choices) > 0 {
		// tmpRes, _ := json.Marshal(chunk)
		// fmt.Println(string(tmpRes))
		content := chunk.Choices[0].Delta.Content
		finishReason := chunk.Choices[0].FinishReason
		reasoningContent := chunk.Choices[0].Delta.ReasoningContent
		if content != "" || reasoningContent != "" || len(chunk.Choices[0].Delta.ToolCalls) > 0 {
			return &model.Response{Content: content, ReasoningContent: reasoningContent, ToolCalls: chunk.Choices[0].Delta.ToolCalls}
		}

		if finishReason == "stop" {
			return nil
		}
	}
	return nil
}

func ParseResponse(resp *http.Response) (*model.Response, error) {
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll failed, err: %v", err)
	}
	defer resp.Body.Close()
	// 解析响应结构
	var chunk struct {
		Choices []struct {
			Delta struct {
				Content          string           `json:"content"`
				ReasoningContent string           `json:"reasoning_content"`
				ToolCalls        []map[string]any `json:"tool_calls"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(all, &chunk); err != nil {
		return &model.Response{Content: string(all)}, nil
	}
	// tmpRes, _ := json.Marshal(chunk)
	// fmt.Println(string(tmpRes))
	if chunk.Choices[0].FinishReason == "tool_calls" {
		return &model.Response{Content: chunk.Choices[0].Delta.Content, ToolCalls: chunk.Choices[0].Delta.ToolCalls}, nil
	}
	return &model.Response{Content: chunk.Choices[0].Delta.Content}, nil
}
