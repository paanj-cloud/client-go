package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ClientHttpClient struct {
	apiKey             string
	apiUrl             string
	accessToken        string
	refreshTokenCallback func() error
	client             *http.Client
}

func NewClientHttpClient(apiKey, apiUrl string) *ClientHttpClient {
	return &ClientHttpClient{
		apiKey: apiKey,
		apiUrl: apiUrl,
		client: &http.Client{},
	}
}

func (c *ClientHttpClient) SetAccessToken(token string) {
	c.accessToken = token
}

func (c *ClientHttpClient) SetRefreshTokenCallback(callback func() error) {
	c.refreshTokenCallback = callback
}

func (c *ClientHttpClient) Request(method, path string, body interface{}, skipAuth bool) (map[string]interface{}, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.apiUrl+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)

	if !skipAuth && c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized && c.refreshTokenCallback != nil && !skipAuth {
		// Token might be expired, try refreshing
		if err := c.refreshTokenCallback(); err == nil {
			// Retry request with new token
			req.Header.Set("Authorization", "Bearer "+c.accessToken)
			// Re-create body reader as it might have been read
			if body != nil {
				jsonBody, _ := json.Marshal(body)
				req.Body = io.NopCloser(bytes.NewBuffer(jsonBody))
			}
			resp, err = c.client.Do(req)
			if err != nil {
				return nil, fmt.Errorf("retry request failed: %w", err)
			}
			defer resp.Body.Close()
		}
	}

	if resp.StatusCode >= 400 {
		responseBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api error: status=%d body=%s", resp.StatusCode, string(responseBody))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		// It's possible the response is empty or not JSON
		return nil, nil 
	}

	return result, nil
}
