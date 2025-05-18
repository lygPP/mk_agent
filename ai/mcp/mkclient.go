package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"lygPP/mk_agent/model"
	"sync"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type MkMcpClient struct {
	client *client.Client
}

var singleMkMcpClient *MkMcpClient
var newClienLock sync.Mutex

func NewMkMcpClient(ctx context.Context) *MkMcpClient {
	if singleMkMcpClient == nil {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("NewMkMcpClient panic:%+v", err)
			}
			newClienLock.Unlock()
		}()
		newClienLock.Lock()
		if singleMkMcpClient == nil {
			mcpClient, err := client.NewStdioMCPClient(
				"/Users/mklin/Desktop/mk_docs/mcp_server/output/mcp_server",
				[]string{},
			)
			if err != nil {
				panic(err)
			}
			initRequest := mcp.InitializeRequest{}
			initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
			initRequest.Params.ClientInfo = mcp.Implementation{
				Name:    "MkMcpClient",
				Version: "1.0.0",
			}
			initResult, err := mcpClient.Initialize(ctx, initRequest)
			if err != nil {
				panic(err)
			}
			fmt.Printf("初始化mcpClient成功，服务器信息: %s %s\n", initResult.ServerInfo.Name, initResult.ServerInfo.Version)
			singleMkMcpClient = &MkMcpClient{
				client: mcpClient,
			}
		}
	}
	return singleMkMcpClient
}

func (c *MkMcpClient) GenDeeepseekTools(ctx context.Context) ([]model.Tool, error) {
	res := make([]model.Tool, 0)
	toolsRequest := mcp.ListToolsRequest{}
	toolResult, err := c.client.ListTools(ctx, toolsRequest)
	if err != nil {
		return nil, err
	}
	if len(toolResult.Tools) == 0 {
		return nil, nil
	}
	for _, tool := range toolResult.Tools {
		tmpResTool := model.Tool{
			Type: "function",
			Function: model.Function{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters: map[string]interface{}{
					"type": "object",
					// "properties": make(map[string]interface{}),
					"required": make([]string, 0),
				},
			},
		}
		propertiesMap := make(map[string]map[string]interface{})
		for pName, pInfo := range tool.InputSchema.Properties {
			jStr, _ := json.Marshal(pInfo)
			pInfoMap := make(map[string]interface{})
			_ = json.Unmarshal(jStr, &pInfoMap)
			propertiesMap[pName] = make(map[string]interface{})
			propertiesMap[pName]["type"] = ""
			propertiesMap[pName]["description"] = ""
			if tmpVl := pInfoMap["type"]; tmpVl != nil {
				propertiesMap[pName]["type"] = tmpVl.(string)
			}
			if tmpVl := pInfoMap["description"]; tmpVl != nil {
				propertiesMap[pName]["description"] = tmpVl.(string)
			}
			if tmpVl := pInfoMap["enum"]; tmpVl != nil {
				propertiesMap[pName]["description"] = fmt.Sprintf("%v，枚举值：%v", propertiesMap[pName]["description"], tmpVl)
			}
		}
		tmpResTool.Function.Parameters["properties"] = propertiesMap
		tmpResTool.Function.Parameters["required"] = tool.InputSchema.Required
		res = append(res, tmpResTool)
	}
	return res, nil
}

func (c *MkMcpClient) CallTool(ctx context.Context, toolName string, toolArguments map[string]any) (string, error) {
	toolRequest := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "tools/call",
		},
	}
	toolRequest.Params.Name = toolName
	toolRequest.Params.Arguments = toolArguments
	result, err := c.client.CallTool(ctx, toolRequest)
	if err != nil {
		return "", err
	}
	return result.Content[0].(mcp.TextContent).Text, nil
}
