package hanlderHttp

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/overiss/vectovm-api/internal/model"
	"github.com/overiss/vectovm-api/internal/repository"
	"github.com/overiss/vectovm-api/internal/server/http/middleware"
	vmservice "github.com/overiss/vectovm-api/internal/service/vm"
)

type VMHandler struct {
	service *vmservice.Service
}

func NewVMHandler(service *vmservice.Service) *VMHandler {
	return &VMHandler{service: service}
}

// Create godoc
// @Summary      Create VM
// @Description  Registers a user VM linked to an owned datanode. SSH login/password are envelope-encrypted with the user's DEK and stored in PostgreSQL.
// @Tags         vm
// @Accept       json
// @Produce      json
// @Param        request body model.CreateVMRequest true "VM parameters"
// @Success      201 {object} model.VMResponse
// @Failure      400 {object} model.ErrorResponse
// @Failure      401 {object} model.ErrorResponse
// @Failure      403 {object} model.ErrorResponse
// @Failure      500 {object} model.ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/vms [post]
func (h *VMHandler) Create(c *gin.Context) {
	oauthUserID, ok := middleware.OAuthUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req model.CreateVMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid request body"})
		return
	}

	resp, err := h.service.Create(c.Request.Context(), oauthUserID, req)
	if err != nil {
		if errors.Is(err, vmservice.ErrDatanodeNotOwned) {
			c.JSON(http.StatusForbidden, model.ErrorResponse{Error: "datanode not found or not owned by user"})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to create vm"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// List godoc
// @Summary      List VMs
// @Description  Returns VMs owned by the authenticated user (credentials are not included).
// @Tags         vm
// @Produce      json
// @Success      200 {array} model.VMResponse
// @Failure      401 {object} model.ErrorResponse
// @Failure      500 {object} model.ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/vms [get]
func (h *VMHandler) List(c *gin.Context) {
	oauthUserID, ok := middleware.OAuthUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "unauthorized"})
		return
	}

	resp, err := h.service.List(c.Request.Context(), oauthUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to list vms"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Get godoc
// @Summary      Get VM
// @Description  Returns VM metadata by name (credentials are not included).
// @Tags         vm
// @Produce      json
// @Param        name path string true "VM name"
// @Success      200 {object} model.VMResponse
// @Failure      401 {object} model.ErrorResponse
// @Failure      404 {object} model.ErrorResponse
// @Failure      500 {object} model.ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/vms/{name} [get]
func (h *VMHandler) Get(c *gin.Context) {
	oauthUserID, ok := middleware.OAuthUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "unauthorized"})
		return
	}

	resp, err := h.service.Get(c.Request.Context(), oauthUserID, c.Param("name"))
	if err != nil {
		if errors.Is(err, repository.ErrVMNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "vm not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get vm"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
