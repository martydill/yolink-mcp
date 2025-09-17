package main

// YoLink API Authentication Response
type YoLinkAuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Code        int    `json:"code"`
	Message     string `json:"message"`
}

// YoLink Device represents a device in the YoLink ecosystem
type YoLinkDevice struct {
	DeviceID       string      `json:"deviceId"`
	DeviceUDID     string      `json:"deviceUDID"`
	DeviceName     string      `json:"name"`
	Token          string      `json:"token"`
	DeviceType     string      `json:"type"`
	ParentDeviceID interface{} `json:"parentDeviceId"`
	ModelName      string      `json:"modelName"`
	ServiceZone    string      `json:"serviceZone"`
}

// YoLink Device List Response
type YoLinkDeviceListResponse struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
	Time    int64  `json:"time"`
	MsgID   int64  `json:"msgid"`
	Method  string `json:"method"`
	Desc    string `json:"desc"`
	Data    struct {
		Devices []YoLinkDevice `json:"devices"`
	} `json:"data"`
}

// YoLink Device Status Response
type YoLinkDeviceStatusResponse struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
	Time    int64  `json:"time"`
	MsgID   int64  `json:"msgid"`
	Method  string `json:"method"`
	Desc    string `json:"desc"`
	Data    map[string]interface{} `json:"data"`
}

// MCP Protocol Types
type MCPRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCP Tool Definitions
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type MCPCapabilities struct {
	Tools struct {
		ListChanged bool `json:"listChanged,omitempty"`
	} `json:"tools,omitempty"`
}

type MCPInitializeResult struct {
	ProtocolVersion string          `json:"protocolVersion"`
	Capabilities    MCPCapabilities `json:"capabilities"`
	ServerInfo      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"serverInfo"`
}

// Tool Call Types
type MCPToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type MCPToolResult struct {
	Content []map[string]interface{} `json:"content"`
	IsError bool                     `json:"isError,omitempty"`
}