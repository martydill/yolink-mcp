package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

func main() {
	// Set up logging
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

	// Check for required environment variables
	clientID := os.Getenv("YOLINK_CLIENT_ID")
	clientSecret := os.Getenv("YOLINK_CLIENT_SECRET")

	// Track missing variables
	var missingVars []string

	if clientID == "" {
		logrus.Error("Environment variables - YOLINK_CLIENT_ID: [NOT SET]")
		missingVars = append(missingVars, "YOLINK_CLIENT_ID")
	}

	if clientSecret == "" {
		logrus.Error("Environment variables - YOLINK_CLIENT_SECRET: [NOT SET]")
		missingVars = append(missingVars, "YOLINK_CLIENT_SECRET")
	}

	// Exit early if required environment variables are missing
	if len(missingVars) > 0 {
		logrus.Errorf("Missing required environment variables: %v", missingVars)
		logrus.Error("Please set the required environment variables and try again.")
		logrus.Error("Example:")
		logrus.Error("  export YOLINK_CLIENT_ID=\"your_client_id_here\"")
		logrus.Error("  export YOLINK_CLIENT_SECRET=\"your_client_secret_here\"")
		logrus.Error("Get your credentials from: https://developer.yosmart.com/")
		os.Exit(1)
	}

	// Create MCP server
	server := NewMCPServer()

	// Set up stdin/stdout communication
	scanner := bufio.NewScanner(os.Stdin)

	logrus.Info("YoLink MCP Server starting...")

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		response, err := server.HandleRequest(line)
		if err != nil {
			logrus.WithError(err).Error("Error handling request")
			continue
		}

		fmt.Println(response)
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		logrus.WithError(err).Fatal("Error reading from stdin")
	}
}
