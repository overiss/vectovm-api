package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	oauthclient "github.com/overiss/vectovm-api/internal/client/oauth"
	cryptosvc "github.com/overiss/vectovm-api/internal/crypto"
	"github.com/overiss/vectovm-api/internal/model"
	"github.com/overiss/vectovm-api/internal/repository"
)

var (
	ErrEmailAlreadyRegistered = errors.New("email already registered")
	ErrInvalidCredentials     = errors.New("invalid request")
)

type Service struct {
	oauth       *oauthclient.Client
	users       *repository.UserRepository
	credentials *cryptosvc.CredentialService
}

func NewService(oauth *oauthclient.Client, users *repository.UserRepository, credentials *cryptosvc.CredentialService) *Service {
	return &Service{
		oauth:       oauth,
		users:       users,
		credentials: credentials,
	}
}

func (s *Service) SignUp(ctx context.Context, req model.SignUpRequest) (*model.SignUpResponse, error) {
	resp, status, err := s.oauth.RegisterUser(ctx, oauthclient.RegisterUserRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if status == http.StatusConflict {
			return nil, ErrEmailAlreadyRegistered
		}
		if status == http.StatusBadRequest {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("register oauth user: %w", err)
	}

	oauthUserID, err := uuid.Parse(resp.UserID)
	if err != nil {
		return nil, fmt.Errorf("parse oauth user id: %w", err)
	}

	encryptedDEK, err := s.credentials.GenerateUserDEK()
	if err != nil {
		return nil, fmt.Errorf("generate user dek: %w", err)
	}

	user, err := s.users.Create(ctx, oauthUserID, encryptedDEK)
	if err != nil {
		return nil, fmt.Errorf("persist user: %w", err)
	}

	return &model.SignUpResponse{
		UserID:      user.ID.String(),
		OAuthUserID: user.OAuthUserID.String(),
	}, nil
}

func (s *Service) ExchangeCode(ctx context.Context, req model.TokenExchangeRequest) (*model.TokenResponse, error) {
	tokens, err := s.oauth.ExchangeCode(ctx, req.Code, req.CodeVerifier)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}
	return mapTokenResponse(tokens), nil
}

func (s *Service) Refresh(ctx context.Context, req model.RefreshTokenRequest) (*model.TokenResponse, error) {
	tokens, err := s.oauth.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("refresh token: %w", err)
	}
	return mapTokenResponse(tokens), nil
}

func (s *Service) Logout(ctx context.Context, req model.LogoutRequest) error {
	if err := s.oauth.RevokeToken(ctx, req.RefreshToken, "refresh_token"); err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}
	return nil
}

func mapTokenResponse(tokens *oauthclient.TokenResponse) *model.TokenResponse {
	return &model.TokenResponse{
		AccessToken:  tokens.AccessToken,
		TokenType:    tokens.TokenType,
		ExpiresIn:    tokens.ExpiresIn,
		RefreshToken: tokens.RefreshToken,
		Scope:        tokens.Scope,
	}
}

func BearerToken(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
