package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL      string
	clientID     string
	clientSecret string
	redirectURI  string
	httpClient   *http.Client
}

func NewClient(baseURL, clientID, clientSecret, redirectURI string) *Client {
	return &Client{
		baseURL:      strings.TrimRight(baseURL, "/"),
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type RegisterUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterUserResponse struct {
	UserID string `json:"user_id"`
}

func (c *Client) RegisterUser(ctx context.Context, req RegisterUserRequest) (*RegisterUserResponse, int, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal register user request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/users", bytes.NewReader(payload))
	if err != nil {
		return nil, 0, fmt.Errorf("build register user request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, 0, fmt.Errorf("register user request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, resp.StatusCode, fmt.Errorf("register user failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var result RegisterUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("decode register user response: %w", err)
	}
	return &result, resp.StatusCode, nil
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

func (c *Client) ExchangeCode(ctx context.Context, code, codeVerifier string) (*TokenResponse, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {c.redirectURI},
		"code_verifier": {codeVerifier},
	}
	return c.tokenRequest(ctx, data)
}

func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}
	return c.tokenRequest(ctx, data)
}

func (c *Client) tokenRequest(ctx context.Context, data url.Values) (*TokenResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.SetBasicAuth(c.clientID, c.clientSecret)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("token request failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var result TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}
	return &result, nil
}

func (c *Client) RevokeToken(ctx context.Context, token, tokenTypeHint string) error {
	data := url.Values{
		"token":           {token},
		"token_type_hint": {tokenTypeHint},
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/oauth/revoke", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("build revoke request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.SetBasicAuth(c.clientID, c.clientSecret)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("revoke request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("revoke failed: status=%d body=%s", resp.StatusCode, string(body))
	}
	return nil
}

type IntrospectResponse struct {
	Active bool `json:"active"`
}

func (c *Client) Introspect(ctx context.Context, token string) (bool, error) {
	payload, err := json.Marshal(map[string]string{"token": token})
	if err != nil {
		return false, fmt.Errorf("marshal introspect request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/tokens/introspect", bytes.NewReader(payload))
	if err != nil {
		return false, fmt.Errorf("build introspect request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.SetBasicAuth(c.clientID, c.clientSecret)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return false, fmt.Errorf("introspect request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return false, fmt.Errorf("introspect failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var result IntrospectResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("decode introspect response: %w", err)
	}
	return result.Active, nil
}

func (c *Client) JWKSURL() string {
	return c.baseURL + "/.well-known/jwks.json"
}
