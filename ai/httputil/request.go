package httputil

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"lygPP/mk_agent/model"
	"net/http"
)

// DoRequest 执行请求
func DoRequest(ctx context.Context, req model.Request, stream bool, c *Config) (*http.Response, error) {
	req.Stream = stream
	body, _ := json.Marshal(req)

	httpReq, _ := http.NewRequestWithContext(
		ctx,
		"POST",
		c.URL,
		bytes.NewReader(body),
	)
	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		fmt.Printf("[DoRequest] err:%+v", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[DoRequest] err:%+v", resp)
		return nil, errors.New("StatusCode != 200")
	}
	return resp, nil
}

func HandleStreaming(ctx context.Context, req model.Request, c *Config) (<-chan model.Response, error) {
	ch := make(chan model.Response)
	go func() {
		defer close(ch)
		resp, err := DoRequest(ctx, req, true, c)
		if err != nil {
			ch <- model.Response{Content: "request error!"}
			return
		}
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				ch <- model.Response{Content: "ctx done!"}
				return
			default:
				event := ParseEvent(scanner.Bytes())
				if event != nil {
					ch <- *event
				}
			}
		}
	}()

	return ch, nil
}
