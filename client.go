package client

import (
	"fmt"
	"time"
)

type ClientOptions struct {
	ApiKey               string
	ApiUrl               string
	WsUrl                string
	AutoReconnect        bool
	ReconnectInterval    time.Duration
	MaxReconnectAttempts int
}

type AuthResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	UserId       string `json:"userId"`
	ExpiresIn    int    `json:"expiresIn"`
}

type PaanjClient struct {
	apiKey     string
	wsClient   *ClientWebSocketClient
	httpClient *ClientHttpClient
	options    ClientOptions
	params     ClientOptions // Normalized options

	accessToken  string
	refreshToken string
	userId       string
}

func NewClient(options ClientOptions) *PaanjClient {
	if options.ApiKey == "" {
		panic("API Key is required")
	}

	params := options
	if params.ApiUrl == "" {
		params.ApiUrl = "http://localhost:3000"
	}
	if params.WsUrl == "" {
		params.WsUrl = "ws://localhost:8090"
	}
	if params.ReconnectInterval == 0 {
		params.ReconnectInterval = 5 * time.Second
	}
	if params.MaxReconnectAttempts == 0 {
		params.MaxReconnectAttempts = 10
	}
	// Default AutoReconnect to true if not specified? Go bool default is false.
	// We might need a pointer for bool or just assume false is default behavior if user didn't set it.
	// For now, let's just use the value provided.

	client := &PaanjClient{
		apiKey:  options.ApiKey,
		options: options,
		params:  params,
	}

	client.wsClient = NewClientWebSocketClient(
		params.ApiKey,
		params.WsUrl,
		params.AutoReconnect,
		params.ReconnectInterval,
		params.MaxReconnectAttempts,
	)

	client.httpClient = NewClientHttpClient(params.ApiKey, params.ApiUrl)

	client.httpClient.SetRefreshTokenCallback(func() error {
		_, err := client.RefreshAccessToken()
		return err
	})

	return client
}

func (c *PaanjClient) AuthenticateAnonymous(userData map[string]interface{}, privateData interface{}) (*AuthResponse, error) {
	body := map[string]interface{}{
		"user":    userData,
		"private": privateData,
	}

	resp, err := c.httpClient.Request("POST", "/api/v1/users/anonymous", body, true)
	if err != nil {
		return nil, err
	}

	// Helper to convert userId which may be float64 or string
	var userId string
	switch v := resp["userId"].(type) {
	case string:
		userId = v
	case float64:
		userId = fmt.Sprintf("%.0f", v)
	default:
		userId = fmt.Sprintf("%v", v)
	}

	authResp := &AuthResponse{
		AccessToken:  resp["accessToken"].(string),
		RefreshToken: resp["refreshToken"].(string),
		UserId:       userId,
		// ExpiresIn might need type assertion care
	}

	c.setSession(authResp)

	c.wsClient.Emit("user.created", map[string]interface{}{
		"userId":       authResp.UserId,
		"accessToken":  authResp.AccessToken,
		"refreshToken": authResp.RefreshToken,
	})

	return authResp, nil
}

func (c *PaanjClient) AuthenticateWithToken(token string, userId string, refreshToken string) {
	c.accessToken = token
	c.refreshToken = refreshToken
	c.userId = userId

	c.wsClient.SetAccessToken(token)
	c.httpClient.SetAccessToken(token)

	c.wsClient.Emit("token.updated", map[string]interface{}{
		"userId":       userId,
		"accessToken":  token,
		"refreshToken": refreshToken,
	})
}

func (c *PaanjClient) setSession(session *AuthResponse) {
	c.accessToken = session.AccessToken
	c.refreshToken = session.RefreshToken
	c.userId = session.UserId

	c.wsClient.SetAccessToken(session.AccessToken)
	c.httpClient.SetAccessToken(session.AccessToken)
}

func (c *PaanjClient) Connect() error {
	return c.wsClient.Connect()
}

func (c *PaanjClient) Disconnect() {
	c.wsClient.Disconnect()
}

func (c *PaanjClient) IsConnected() bool {
	return c.wsClient.IsConnectedStatus()
}

func (c *PaanjClient) IsAuthenticated() bool {
	return c.accessToken != ""
}

func (c *PaanjClient) GetUserId() string {
	return c.userId
}

func (c *PaanjClient) RefreshAccessToken() (*AuthResponse, error) {
	// Implement refresh logic... simplified for now
	return nil, nil
}

// Internal Accessors

func (c *PaanjClient) GetWebSocket() *ClientWebSocketClient {
	return c.wsClient
}

func (c *PaanjClient) GetHttpClient() *ClientHttpClient {
	return c.httpClient
}

func (c *PaanjClient) Subscribe(subscription interface{}) error {
	return c.wsClient.Subscribe(subscription)
}

func (c *PaanjClient) On(event string, callback func(interface{})) {
	c.wsClient.On(event, callback)
}
