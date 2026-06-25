package model

import "time"

type ErrorResponse struct {
	Error string `json:"error" example:"invalid request"`
}

type HealthResponse struct {
	Status  string `json:"status" example:"ok"`
	Version string `json:"version" example:"v1.0.0"`
}

type ReadyResponse struct {
	Status string `json:"status" example:"ready"`
}

type SignUpRequest struct {
	Email    string `json:"email" binding:"required,email" example:"alice@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"her-secure-password"`
}

type SignUpResponse struct {
	UserID      string `json:"user_id" example:"6ba7b810-9dad-11d1-80b4-00c04fd430c8"`
	OAuthUserID string `json:"oauth_user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type TokenExchangeRequest struct {
	Code         string `json:"code" binding:"required"`
	CodeVerifier string `json:"code_verifier" binding:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UserResponse struct {
	ID          string    `json:"id"`
	OAuthUserID string    `json:"oauth_user_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateDatanodeRequest struct {
	Name     string `json:"name" binding:"required" example:"primary-datanode"`
	Host     string `json:"host" binding:"required" example:"10.0.0.5"`
	Port     int    `json:"port" example:"22"`
	User     string `json:"user" binding:"required" example:"root"`
	Password string `json:"password" binding:"required" example:"secret"`
}

type DatanodeResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	SSHUser   string    `json:"ssh_user"`
	LastJobID *string   `json:"last_job_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type DeployVaultRequest struct {
	DatanodeName string `json:"datanode_name" binding:"required" example:"primary-datanode"`
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
	Error      *string `json:"error,omitempty"`
	CreatedAt  string  `json:"created_at"`
	FinishedAt *string `json:"finished_at,omitempty"`
}

type RuntimeResponse struct {
	UserID       string `json:"user_id"`
	DatanodeName string `json:"datanode_name"`
	Datanode     string `json:"datanode"`
	VaultStatus  string `json:"vault_status"`
	VaultLogs    string `json:"vault_logs"`
}

type CreateVMRequest struct {
	Name         string `json:"name" binding:"required" example:"my-app-vm"`
	DatanodeName string `json:"datanode_name" binding:"required" example:"primary-datanode"`
	Host         string `json:"host" binding:"required" example:"10.0.0.10"`
	Port         int    `json:"port" example:"22"`
	SSHUser      string `json:"ssh_user" binding:"required" example:"ubuntu"`
	SSHPassword  string `json:"ssh_password" binding:"required" example:"secret"`
}

type VMResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	DatanodeName string    `json:"datanode_name"`
	Host         string    `json:"host"`
	Port         int       `json:"port"`
	CreatedAt    time.Time `json:"created_at"`
}
