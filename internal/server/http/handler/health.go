package hanlderHttp

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/overiss/vectovm-api/internal/model"
	"github.com/overiss/vectovm-api/internal/storage/postgres"
)

type HealthHandler struct {
	version string
	db      *postgres.DB
}

func NewHealthHandler(version string, db *postgres.DB) *HealthHandler {
	return &HealthHandler{
		version: version,
		db:      db,
	}
}

// Healthz godoc
// @Summary      Liveness probe
// @Tags         health
// @Produce      json
// @Success      200 {object} model.HealthResponse
// @Router       /healthz [get]
func (h *HealthHandler) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, model.HealthResponse{
		Status:  "ok",
		Version: h.version,
	})
}

// Readyz godoc
// @Summary      Readiness probe
// @Tags         health
// @Produce      json
// @Success      200 {object} model.ReadyResponse
// @Failure      503 {object} model.ErrorResponse
// @Router       /readyz [get]
func (h *HealthHandler) Readyz(c *gin.Context) {
	if err := h.db.Ping(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, model.ErrorResponse{Error: "postgres unavailable"})
		return
	}
	c.JSON(http.StatusOK, model.ReadyResponse{Status: "ready"})
}

func (h *HealthHandler) Ping(ctx context.Context) error {
	return h.db.Ping(ctx)
}
