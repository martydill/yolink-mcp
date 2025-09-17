# YoLink MCP Server

Finally - the security of IOT devices combined with the security of MCP!

<img width="1389" height="792" alt="Screenshot 2025-09-16 113015" src="https://github.com/user-attachments/assets/7c2b4431-573e-464c-9dac-fe07c7806fb5" />

This is a Model Context Protocol (MCP) server for interacting with YoLink smart devices. This server allows you to enumerate devices and get status information from your YoLink account through the MCP protocol.



## Features

- **Device Enumeration**: List all devices in your YoLink account
- **Device Status**: Get current status and state information for any device
- **MCP Protocol**: Full compliance with MCP specification for seamless integration
- **Authentication**: Secure OAuth2 authentication with YoLink API
- **Error Handling**: Comprehensive error handling and logging

## Prerequisites

- Go 1.21 or later
- YoLink account with API access
- YoLink Client ID and Client Secret

## Installation

1. Clone this repository:
```bash
git clone <repository-url>
cd yolink-mcp
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the server:
```bash
go build -o yolink-mcp-server .
```

## Configuration

Set the following environment variables with your YoLink API credentials:

```bash
export YOLINK_CLIENT_ID="your_client_id_here"
export YOLINK_CLIENT_SECRET="your_client_secret_here"
```

### Getting YoLink API Credentials

1. Visit the [YoLink Developer Portal](https://developer.yosmart.com/)
2. Create an account or sign in
3. Create a new application to get your Client ID and Client Secret
4. Make sure your application has the `yolink:read` scope

## Usage

### Running the Server

Start the MCP server:

```bash
./yolink-mcp-server
```

The server will listen for MCP requests on stdin and send responses to stdout.

### Available Tools

#### 1. enumerate_devices
Lists all YoLink devices in your account.

**Parameters**: None

**Example MCP Request**:
```json
{
  "jsonrpc": "2.0",
  "id": "1",
  "method": "tools/call",
  "params": {
    "name": "enumerate_devices",
    "arguments": {}
  }
}
```

#### 2. get_device_status
Gets the current status of a specific YoLink device.

**Parameters**:
- `device_id` (string): The ID of the device to get status for

**Example MCP Request**:
```json
{
  "jsonrpc": "2.0",
  "id": "2",
  "method": "tools/call",
  "params": {
    "name": "get_device_status",
    "arguments": {
      "device_id": "your_device_id_here"
    }
  }
}
```

### MCP Integration

This server implements the MCP protocol and can be integrated with any MCP-compatible client. The server supports:

- `initialize`: Server initialization and capability advertisement
- `tools/list`: List available tools
- `tools/call`: Execute tool calls

### Example Integration

You can integrate this server with MCP clients by configuring the client to use this server as a tool provider. The server communicates via stdin/stdout using JSON-RPC messages.

## Project Structure

```
yolink-mcp/
├── main.go           # Main entry point and stdin/stdout handling
├── mcp_server.go     # MCP protocol implementation
├── yolink_client.go  # YoLink API client
├── types.go          # Data structures and type definitions
├── go.mod            # Go module file
└── README.md         # This file
```

## Error Handling

The server includes comprehensive error handling for:

- Authentication failures
- Network connectivity issues
- Invalid device IDs
- API rate limiting
- Malformed MCP requests

All errors are logged using structured logging and returned as proper MCP error responses.

## Development

### Building from Source

```bash
go build -o yolink-mcp-server .
```

### Running Tests

```bash
go test ./...
```

### Logging

The server uses structured logging with the following levels:
- `INFO`: General operational information
- `ERROR`: Error conditions and failures
- `DEBUG`: Detailed debugging information (when enabled)

## API Reference

### YoLink API Endpoints Used

- `POST /oauth/token`: Authentication
- `POST /openapi`: Device operations (getDeviceList, getState)

### Supported Device Types

This server works with all YoLink device types including:
- Smart switches and outlets
- Sensors (temperature, humidity, door/window, motion, etc.)
- Smart locks
- Garage door controllers
- Water leak sensors
- And more...

## Troubleshooting

### Common Issues

1. **Authentication Error**: Verify your `YOLINK_CLIENT_ID` and `YOLINK_CLIENT_SECRET` are correct
2. **Network Issues**: Check your internet connection and firewall settings
3. **Device Not Found**: Ensure the device ID is correct and the device is online
4. **Permission Denied**: Verify your YoLink application has the correct scopes

### Debug Mode

Set log level to debug for verbose output:
```bash
export LOG_LEVEL=debug
./yolink-mcp-server
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- YoLink for providing the IoT device API
 The MCP protocol specification and community
