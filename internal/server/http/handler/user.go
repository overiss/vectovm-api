package hanlderHttp

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/overiss/vectovm-api/internal/model"
	"github.com/overiss/vectovm-api/internal/repository"
	"github.com/overiss/vectovm-api/internal/server/http/middleware"
	datanodeservice "github.com/overiss/vectovm-api/internal/service/datanode"
	userservice "github.com/overiss/vectovm-api/internal/service/user"
)

type UserHandler struct {
	service *userservice.Service
}

func NewUserHandler(service *userservice.Service) *UserHandler {
	return &UserHandler{service: service}
}

// Me godoc
// @Summary      Current user profile
// @Description  Returns domain user record linked to the OAuth subject.
// @Tags         user
// @Produce      json
// @Success      200 {object} model.UserResponse
// @Failure      401 {object} model.ErrorResponse
// @Failure      404 {object} model.ErrorResponse
// @Failure      500 {object} model.ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/me [get]
func (h *UserHandler) Me(c *gin.Context) {
	oauthUserID, ok := middleware.OAuthUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "unauthorized"})
		return
	}

	resp, err := h.service.GetByOAuthUserID(c.Request.Context(), oauthUserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to load user"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

type DatanodeHandler struct {
	service *datanodeservice.Service
}

func NewDatanodeHandler(service *datanodeservice.Service) *DatanodeHandler {
	return &DatanodeHandler{service: service}
}

// Create godoc
// @Summary      Create datanode
// @Description  Provisions a datanode via vectovm-mapi (async). SSH credentials are stored in mapi-vault.
// @Tags         datanode
// @Accept       json
// @Produce      json
// @Param        request body model.CreateDatanodeRequest true "Datanode SSH credentials"
// @Success      202 {object} model.JobResponse
// @Failure      400 {object} model.ErrorResponse
// @Failure      401 {object} model.ErrorResponse
// @Failure      502 {object} model.ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/datanodes [post]
func (h *DatanodeHandler) Create(c *gin.Context) {
	oauthUserID, ok := middleware.OAuthUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req model.CreateDatanodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid request body"})
		return
	}

	resp, err := h.service.Create(c.Request.Context(), oauthUserID, req)
	if err != nil {
		c.JSON(http.StatusBadGateway, model.ErrorResponse{Error: "failed to create datanode"})
		return
	}

	c.JSON(http.StatusAccepted, resp)
}

// List godoc
// @Summary      List datanodes
// @Description  Returns datanodes owned by the authenticated user.
// @Tags         datanode
// @Produce      json
// @Success      200 {array} model.DatanodeResponse
// @Failure      401 {object} model.ErrorResponse
// @Failure      500 {object} model.ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/datanodes [get]
func (h *DatanodeHandler) List(c *gin.Context) {
	oauthUserID, ok := middleware.OAuthUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "unauthorized"})
		return
	}

	resp, err := h.service.List(c.Request.Context(), oauthUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to list datanodes"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeployVault godoc
// @Summary      Deploy Vault on datanode
// @Description  Starts async Vault deployment on an existing datanode via vectovm-mapi.
// @Tags         datanode
// @Accept       json
// @Produce      json
// @Param        request body model.DeployVaultRequest true "Target datanode"
// @Success      202 {object} model.JobResponse
// @Failure      400 {object} model.ErrorResponse
// @Failure      401 {object} model.ErrorResponse
// @Failure      404 {object} model.ErrorResponse
// @Failure      502 {object} model.ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/datanodes/vault/deploy [post]
func (h *DatanodeHandler) DeployVault(c *gin.Context) {
	oauthUserID, ok := middleware.OAuthUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req model.DeployVaultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid request body"})
		return
	}

	resp, err := h.service.DeployVault(c.Request.Context(), oauthUserID, req)
	if err != nil {
		if errors.Is(err, repository.ErrDatanodeNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "datanode not found"})
			return
		}
		c.JSON(http.StatusBadGateway, model.ErrorResponse{Error: "failed to deploy vault"})
		return
	}

	c.JSON(http.StatusAccepted, resp)
}

// JobStatus godoc
// @Summary      Get provisioning job status
// @Description  Polls async job status from vectovm-mapi.
// @Tags         datanode
// @Produce      json
// @Param        id path string true "Job ID"
// @Success      200 {object} model.JobStatusResponse
// @Failure      401 {object} model.ErrorResponse
// @Failure      502 {object} model.ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/datanodes/jobs/{id} [get]
func (h *DatanodeHandler) JobStatus(c *gin.Context) {
	oauthUserID, ok := middleware.OAuthUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "unauthorized"})
		return
	}

	resp, err := h.service.GetJob(c.Request.Context(), oauthUserID, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadGateway, model.ErrorResponse{Error: "failed to get job status"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Runtime godoc
// @Summary      Get datanode runtime info
// @Description  Fetches live datanode and Vault status over SSH via vectovm-mapi.
// @Tags         datanode
// @Produce      json
// @Param        name path string true "Datanode name"
// @Success      200 {object} model.RuntimeResponse
// @Failure      400 {object} model.ErrorResponse
// @Failure      401 {object} model.ErrorResponse
// @Failure      404 {object} model.ErrorResponse
// @Failure      502 {object} model.ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/datanodes/{name}/runtime [get]
func (h *DatanodeHandler) Runtime(c *gin.Context) {
	oauthUserID, ok := middleware.OAuthUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "unauthorized"})
		return
	}

	datanodeName := c.Param("name")
	if datanodeName == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "datanode name is required"})
		return
	}

	resp, err := h.service.GetRuntime(c.Request.Context(), oauthUserID, datanodeName)
	if err != nil {
		if errors.Is(err, repository.ErrDatanodeNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "datanode not found"})
			return
		}
		c.JSON(http.StatusBadGateway, model.ErrorResponse{Error: "failed to get runtime info"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
