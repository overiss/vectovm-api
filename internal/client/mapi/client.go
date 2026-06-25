package mapi

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/overiss/vectovm-api/internal/config"
)

type Client struct {
	baseURL     string
	bearerToken string
	httpClient  *http.Client
}

func NewClient(cfg config.Mapi) (*Client, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	if cfg.CACertFile != "" {
		caCert, err := os.ReadFile(cfg.CACertFile)
		if err != nil {
			return nil, fmt.Errorf("read mapi ca cert: %w", err)
		}

		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("parse mapi ca cert")
		}

		transport.TLSClientConfig = &tls.Config{
			RootCAs:    pool,
			MinVersion: tls.VersionTLS12,
		}
	}

	return &Client{
		baseURL:     strings.TrimRight(cfg.BaseURL, "/"),
		bearerToken: cfg.BearerToken,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}, nil
}

type CreateDatanodeRequest struct {
	UserID   string `json:"user_id"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type DeployVaultRequest struct {
	UserID       string `json:"user_id"`
	DatanodeName string `json:"datanode_name"`
}

type JobResponse struct {
	JobID   string `json:"job_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type JobStatusResponse struct {
	Job *Job `json:"job"`
}

type Job struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"`
	Status     string  `json:"status"`
	Error      *string `json:"error"`
	CreatedAt  string  `json:"created_at"`
	FinishedAt *string `json:"finished_at"`
}

type RuntimeResponse struct {
	UserID       string `json:"user_id"`
	DatanodeName string `json:"datanode_name"`
	Datanode     string `json:"datanode"`
	VaultStatus  string `json:"vault_status"`
	VaultLogs    string `json:"vault_logs"`
}

func (c *Client) CreateDatanode(ctx context.Context, req CreateDatanodeRequest) (*JobResponse, error) {
	return c.postJob(ctx, "/core/datanode/create", req)
}

func (c *Client) DeployVault(ctx context.Context, req DeployVaultRequest) (*JobResponse, error) {
	return c.postJob(ctx, "/core/datanode/vault/deploy", req)
}

func (c *Client) GetJob(ctx context.Context, jobID string) (*JobStatusResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/core/datanode/jobs/"+jobID, nil)
	if err != nil {
		return nil, fmt.Errorf("build get job request: %w", err)
	}
	c.setAuth(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("get job request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.readError(resp)
	}

	var result JobStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode job response: %w", err)
	}
	return &result, nil
}

func (c *Client) GetRuntime(ctx context.Context, userID, datanodeName string) (*RuntimeResponse, error) {
	url := fmt.Sprintf("%s/core/datanode/runtime?user_id=%s&datanode_name=%s",
		c.baseURL, userID, datanodeName)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build runtime request: %w", err)
	}
	c.setAuth(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("runtime request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.readError(resp)
	}

	var result RuntimeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode runtime response: %w", err)
	}
	return &result, nil
}

func (c *Client) postJob(ctx context.Context, path string, payload any) (*JobResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal mapi request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build mapi request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuth(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("mapi request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return nil, c.readError(resp)
	}

	var result JobResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode mapi response: %w", err)
	}
	return &result, nil
}

func (c *Client) setAuth(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.bearerToken)
}

func (c *Client) readError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return fmt.Errorf("mapi request failed: status=%d body=%s", resp.StatusCode, string(body))
}
