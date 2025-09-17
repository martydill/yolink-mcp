package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	YoLinkAPIBaseURL = "https://api.yosmart.com/open/yolink/v2/api"
	YoLinkTokenURL   = "https://api.yosmart.com/open/yolink/token"
)

type YoLinkClient struct {
	clientID     string
	clientSecret string
	accessToken  string
	tokenExpiry  time.Time
	httpClient   *http.Client
}

func NewYoLinkClient() (*YoLinkClient, error) {
	clientID := os.Getenv("YOLINK_CLIENT_ID")
	clientSecret := os.Getenv("YOLINK_CLIENT_SECRET")

	if clientID != "" {
		logrus.Infof("Creating YoLink client with ClientID: [SET] (len: %d)", len(clientID))
	} else {
		logrus.Info("Creating YoLink client with ClientID: [NOT SET]")
	}
	if clientSecret != "" {
		logrus.Infof("ClientSecret: [SET] (len: %d)", len(clientSecret))
	} else {
		logrus.Info("ClientSecret: [NOT SET]")
	}

	if clientID == "" || clientSecret == "" {
		logrus.Error("Missing required environment variables")
		return nil, fmt.Errorf("YOLINK_CLIENT_ID and YOLINK_CLIENT_SECRET environment variables are required")
	}

	logrus.Info("YoLink client created successfully")
	return &YoLinkClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (c *YoLinkClient) authenticate() error {
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		return nil // Token is still valid
	}

	logrus.Info("Authenticating with YoLink API...")

	payload := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     c.clientID,
		"client_secret": c.clientSecret,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal auth payload: %w", err)
	}

	req, err := http.NewRequest("POST", YoLinkTokenURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logrus.WithField("status_code", resp.StatusCode).WithField("response_body", string(body)).Error("Authentication failed")
		return fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	var authResp YoLinkAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	c.accessToken = authResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second)

	logrus.Info("Successfully authenticated with YoLink API")
	return nil
}

func (c *YoLinkClient) makeRequest(payload map[string]interface{}) (*http.Response, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}

	requestURL := YoLinkAPIBaseURL

	var reqBody io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request payload: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest("POST", requestURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

func (c *YoLinkClient) GetDevices() ([]YoLinkDevice, error) {
	logrus.Info("Fetching device list from YoLink API")

	payload := map[string]interface{}{
		"method":    "Home.getDeviceList",
		"timestamp": time.Now().Unix(),
	}

	resp, err := c.makeRequest(payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logrus.WithField("status_code", resp.StatusCode).WithField("method", "Home.getDeviceList").WithField("response_body", string(body)).Error("API request failed")
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var deviceResp YoLinkDeviceListResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceResp); err != nil {
		return nil, fmt.Errorf("failed to decode device list response: %w", err)
	}

	if deviceResp.Code != "000000" {
		return nil, fmt.Errorf("API error: %s - %s", deviceResp.Code, deviceResp.Message)
	}

	logrus.Infof("Successfully fetched %d devices", len(deviceResp.Data.Devices))
	return deviceResp.Data.Devices, nil
}

func (c *YoLinkClient) GetDeviceStatus(deviceID string) (map[string]interface{}, error) {
	logrus.Infof("Fetching status for device %s", deviceID)

	// First get the device list to get the device's token and type
	devices, err := c.GetDevices()
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	// Find the specific device
	var deviceToken string
	var deviceType string
	for _, device := range devices {
		if device.DeviceID == deviceID {
			deviceToken = device.Token
			deviceType = device.DeviceType
			break
		}
	}

	if deviceToken == "" {
		return nil, fmt.Errorf("device not found or device token not available")
	}

	// Create the request payload with device type specific method
	payload := map[string]interface{}{
		"method":       deviceType + ".getState",
		"targetDevice": deviceID,
		"token":        deviceToken,
		"timestamp":    time.Now().Unix(),
	}

	resp, err := c.makeRequest(payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logrus.WithField("status_code", resp.StatusCode).WithField("device_id", deviceID).WithField("response_body", string(body)).Error("Device status API request failed")
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var statusResp YoLinkDeviceStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return nil, fmt.Errorf("failed to decode device status response: %w", err)
	}

	if statusResp.Code != "000000" {
		return nil, fmt.Errorf("API error: %s - %s", statusResp.Code, statusResp.Desc)
	}

	logrus.Infof("Successfully fetched status for device %s", deviceID)
	return statusResp.Data, nil
}
