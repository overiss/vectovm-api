package hanlderHttp

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/overiss/vectovm-api/internal/model"
	authservice "github.com/overiss/vectovm-api/internal/service/auth"
)

type AuthHandler struct {
	service *authservice.Service
}

func NewAuthHandler(service *authservice.Service) *AuthHandler {
	return &AuthHandler{service: service}
}

// SignUp godoc
// @Summary      Register user
// @Description  Proxies registration to go-oauthv2 and stores oauth_user_id with a per-user encryption DEK in PostgreSQL.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body model.SignUpRequest true "User credentials"
// @Success      201 {object} model.SignUpResponse
// @Failure      400 {object} model.ErrorResponse
// @Failure      409 {object} model.ErrorResponse
// @Failure      502 {object} model.ErrorResponse
// @Router       /api/v1/signup [post]
func (h *AuthHandler) SignUp(c *gin.Context) {
	var req model.SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid request body"})
		return
	}

	resp, err := h.service.SignUp(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, authservice.ErrEmailAlreadyRegistered):
			c.JSON(http.StatusConflict, model.ErrorResponse{Error: "email already registered"})
		case errors.Is(err, authservice.ErrInvalidCredentials):
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid credentials"})
		default:
			c.JSON(http.StatusBadGateway, model.ErrorResponse{Error: "registration failed"})
		}
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// ExchangeToken godoc
// @Summary      Exchange authorization code for tokens
// @Description  Exchanges OAuth authorization code + PKCE verifier for access and refresh tokens via go-oauthv2.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body model.TokenExchangeRequest true "Authorization code exchange"
// @Success      200 {object} model.TokenResponse
// @Failure      400 {object} model.ErrorResponse
// @Failure      502 {object} model.ErrorResponse
// @Router       /api/v1/auth/token [post]
func (h *AuthHandler) ExchangeToken(c *gin.Context) {
	var req model.TokenExchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid request body"})
		return
	}

	resp, err := h.service.ExchangeCode(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadGateway, model.ErrorResponse{Error: "token exchange failed"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Refresh godoc
// @Summary      Refresh access token
// @Description  Refreshes OAuth session using a refresh token via go-oauthv2.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body model.RefreshTokenRequest true "Refresh token"
// @Success      200 {object} model.TokenResponse
// @Failure      400 {object} model.ErrorResponse
// @Failure      502 {object} model.ErrorResponse
// @Router       /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid request body"})
		return
	}

	resp, err := h.service.Refresh(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadGateway, model.ErrorResponse{Error: "refresh failed"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Logout godoc
// @Summary      Logout
// @Description  Revokes refresh token at go-oauthv2.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body model.LogoutRequest true "Refresh token to revoke"
// @Success      204
// @Failure      400 {object} model.ErrorResponse
// @Failure      502 {object} model.ErrorResponse
// @Router       /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req model.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid request body"})
		return
	}

	if err := h.service.Logout(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusBadGateway, model.ErrorResponse{Error: "logout failed"})
		return
	}

	c.Status(http.StatusNoContent)
}
