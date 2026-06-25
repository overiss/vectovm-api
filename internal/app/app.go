package app

import (
	"context"
	"log"

	"github.com/overiss/vectovm-api/internal/behavior"
	"github.com/overiss/vectovm-api/internal/config"
	"github.com/overiss/vectovm-api/internal/server"
	hanlderHttp "github.com/overiss/vectovm-api/internal/server/http/handler"
	"github.com/overiss/vectovm-api/internal/service"
	"github.com/overiss/vectovm-api/internal/storage/postgres"
)

type Application struct {
	cfg              *config.Application
	netContainer     *server.Container
	handlerContainer *hanlderHttp.Container
	serviceContainer *service.Container
	starter          []behavior.Starter
	readiness        []behavior.Readiness
}

func Init(ctx context.Context, cfg *config.Application) *Application {
	app := &Application{
		cfg:       cfg,
		starter:   make([]behavior.Starter, 0),
		readiness: make([]behavior.Readiness, 0),
	}

	app.initNetContainer()
	app.initServiceContainer(ctx)
	app.initHandlerContainer(ctx)
	app.expose(ctx)

	log.Printf("application initialized successfully")
	return app
}

func (a *Application) Start(ctx context.Context) {
	for _, starter := range a.starter {
		go starter.Start(ctx)
	}
}

func (a *Application) Stop(ctx context.Context) {
	for _, starter := range a.starter {
		starter.Stop(ctx)
	}
	if a.serviceContainer != nil {
		a.serviceContainer.Close()
	}
}

type postgresReadiness struct {
	db *postgres.DB
}

func newPostgresReadiness(db *postgres.DB) *postgresReadiness {
	return &postgresReadiness{db: db}
}

func (p *postgresReadiness) Name() string {
	return "postgres"
}

func (p *postgresReadiness) IsReady() bool {
	return p.db.Ping(context.Background()) == nil
}
