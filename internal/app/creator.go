package app

import (
	"context"
	"log"

	"github.com/overiss/vectovm-api/internal/server"
	serverHttp "github.com/overiss/vectovm-api/internal/server/http"
	hanlderHttp "github.com/overiss/vectovm-api/internal/server/http/handler"
	"github.com/overiss/vectovm-api/internal/service"
)

func (a *Application) initNetContainer() {
	httpServer := serverHttp.NewHttp(a.cfg.Http, a.cfg.IsDebug)

	a.netContainer = server.NewContainer(httpServer)
	a.starter = append(a.starter, httpServer)
}

func (a *Application) initServiceContainer(ctx context.Context) {
	services, err := service.NewContainer(ctx, a.cfg)
	if err != nil {
		log.Fatalf("cannot initialize services: %v", err)
	}

	a.serviceContainer = services
	a.readiness = append(a.readiness, newPostgresReadiness(services.DB()))
}

func (a *Application) initHandlerContainer(ctx context.Context) {
	if a.serviceContainer == nil {
		log.Printf("serviceContainer is not initialized. use initServiceContainer before initHandlerContainer")
		return
	}

	a.handlerContainer = hanlderHttp.NewContainer(
		hanlderHttp.NewAuthHandler(a.serviceContainer.Auth),
		hanlderHttp.NewUserHandler(a.serviceContainer.User),
		hanlderHttp.NewDatanodeHandler(a.serviceContainer.Datanode),
		hanlderHttp.NewVMHandler(a.serviceContainer.VM),
		a.serviceContainer.Verifier,
	)
}
