package main

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

type MCPServer struct {
	yolinkClient *YoLinkClient
	tools        []MCPTool
}

func NewMCPServer() *MCPServer {
	yolinkClient, err := NewYoLinkClient()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create YoLink client")
	}

	server := &MCPServer{
		yolinkClient: yolinkClient,
	}

	server.initializeTools()
	return server
}

func (s *MCPServer) initializeTools() {
	s.tools = []MCPTool{
		{
			Name:        "enumerate_devices",
			Description: "List all YoLink devices in the account",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
				"required":   []string{},
			},
		},
		{
			Name:        "get_device_status",
			Description: "Get the current status of a specific YoLink device",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"device_id": map[string]interface{}{
						"type":        "string",
						"description": "The ID of the device to get status for",
					},
				},
				"required": []string{"device_id"},
			},
		},
	}
}

func (s *MCPServer) HandleRequest(requestStr string) (string, error) {
	var request MCPRequest
	if err := json.Unmarshal([]byte(requestStr), &request); err != nil {
		return s.createErrorResponse(nil, -32700, "Parse error")
	}

	switch request.Method {
	case "initialize":
		return s.handleInitialize(request)
	case "tools/list":
		return s.handleToolsList(request)
	case "tools/call":
		return s.handleToolCall(request)
	default:
		return s.createErrorResponse(request.ID, -32601, "Method not found")
	}
}

func (s *MCPServer) handleInitialize(request MCPRequest) (string, error) {
	result := MCPInitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: MCPCapabilities{
			Tools: struct {
				ListChanged bool `json:"listChanged,omitempty"`
			}{
				ListChanged: false,
			},
		},
		ServerInfo: struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}{
			Name:    "yolink-mcp-server",
			Version: "1.0.0",
		},
	}

	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return s.createErrorResponse(request.ID, -32603, "Internal error")
	}

	return string(responseBytes), nil
}

func (s *MCPServer) handleToolsList(request MCPRequest) (string, error) {
	result := map[string]interface{}{
		"tools": s.tools,
	}

	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return s.createErrorResponse(request.ID, -32603, "Internal error")
	}

	return string(responseBytes), nil
}

func (s *MCPServer) handleToolCall(request MCPRequest) (string, error) {
	// Extract tool call parameters
	nameInterface, ok := request.Params["name"]
	if !ok {
		return s.createErrorResponse(request.ID, -32602, "Missing 'name' parameter")
	}

	toolName, ok := nameInterface.(string)
	if !ok {
		return s.createErrorResponse(request.ID, -32602, "Invalid 'name' parameter type")
	}

	argumentsInterface := request.Params["arguments"]
	var arguments map[string]interface{}
	if argumentsInterface != nil {
		var ok bool
		arguments, ok = argumentsInterface.(map[string]interface{})
		if !ok {
			return s.createErrorResponse(request.ID, -32602, "Invalid 'arguments' parameter type")
		}
	}

	var result MCPToolResult
	var err error

	switch toolName {
	case "enumerate_devices":
		result, err = s.enumerateDevices(arguments)
	case "get_device_status":
		result, err = s.getDeviceStatus(arguments)
	default:
		return s.createErrorResponse(request.ID, -32601, "Unknown tool")
	}

	if err != nil {
		logrus.WithError(err).WithField("tool", toolName).Error("Tool execution failed")
		return s.createErrorResponse(request.ID, -32603, fmt.Sprintf("Tool execution failed: %s", err.Error()))
	}

	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result:  result,
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return s.createErrorResponse(request.ID, -32603, "Internal error")
	}

	return string(responseBytes), nil
}

func (s *MCPServer) enumerateDevices(arguments map[string]interface{}) (MCPToolResult, error) {
	devices, err := s.yolinkClient.GetDevices()
	if err != nil {
		return MCPToolResult{
			Content: []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("Failed to enumerate devices: %s", err.Error()),
				},
			},
			IsError: true,
		}, nil
	}

	devicesJSON, err := json.MarshalIndent(devices, "", "  ")
	if err != nil {
		return MCPToolResult{
			Content: []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("Failed to serialize devices: %s", err.Error()),
				},
			},
			IsError: true,
		}, nil
	}

	return MCPToolResult{
		Content: []map[string]interface{}{
			{
				"type": "text",
				"text": fmt.Sprintf("Found %d YoLink devices:\n\n%s", len(devices), string(devicesJSON)),
			},
		},
	}, nil
}

func (s *MCPServer) getDeviceStatus(arguments map[string]interface{}) (MCPToolResult, error) {
	deviceIDInterface, ok := arguments["device_id"]
	if !ok {
		return MCPToolResult{
			Content: []map[string]interface{}{
				{
					"type": "text",
					"text": "Missing required parameter: device_id",
				},
			},
			IsError: true,
		}, nil
	}

	deviceID, ok := deviceIDInterface.(string)
	if !ok {
		return MCPToolResult{
			Content: []map[string]interface{}{
				{
					"type": "text",
					"text": "device_id must be a string",
				},
			},
			IsError: true,
		}, nil
	}

	status, err := s.yolinkClient.GetDeviceStatus(deviceID)
	if err != nil {
		return MCPToolResult{
			Content: []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("Failed to get device status: %s", err.Error()),
				},
			},
			IsError: true,
		}, nil
	}

	statusJSON, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return MCPToolResult{
			Content: []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("Failed to serialize device status: %s", err.Error()),
				},
			},
			IsError: true,
		}, nil
	}

	return MCPToolResult{
		Content: []map[string]interface{}{
			{
				"type": "text",
				"text": fmt.Sprintf("Status for device %s:\n\n%s", deviceID, string(statusJSON)),
			},
		},
	}, nil
}

func (s *MCPServer) createErrorResponse(id interface{}, code int, message string) (string, error) {
	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal error response: %w", err)
	}

	return string(responseBytes), nil
}
