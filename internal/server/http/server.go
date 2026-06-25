package serverHttp

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/overiss/vectovm-api/internal/config"
)

type Http struct {
	server *http.Server
	router *gin.Engine
	name   string
}

func NewHttp(cfg *config.HttpServer, isDebug bool) *Http {
	if isDebug {
		gin.SetMode(gin.DebugMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())
	gin.SetMode(gin.ReleaseMode)

	server := &http.Server{
		Addr:         cfg.Port,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return &Http{
		name:   cfg.Name,
		server: server,
		router: router,
	}
}

func (s *Http) Start(_ context.Context) {
	log.Printf("%s starting", s.name)

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("%s failed to start: %v", s.name, err)
	}
}

func (s *Http) Stop(ctx context.Context) {
	log.Printf("shutting down server %s", s.name)

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		log.Printf("error shutting down %s: %v", s.name, err)
	}
	log.Printf("%s stopped successfully", s.name)
}

func (s *Http) Router() *gin.Engine {
	return s.router
}
