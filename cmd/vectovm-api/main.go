// @title           VectoVM API
// @version         1.0
// @description     BFF / domain API for the VectoVM platform. Proxies user registration to go-oauthv2, stores domain data in PostgreSQL, provisions datanodes via vectovm-mapi, and manages user VMs with envelope-encrypted SSH credentials.
// @description     Protected routes require a Bearer access token from go-oauthv2.
// @termsOfService  https://overiss.github.io/terms

// @contact.name   VectoVM Team
// @contact.url    https://github.com/overiss/vectovm-api

// @license.name  Proprietary

// @host      localhost:8081
// @BasePath  /
// @schemes   http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description OAuth 2.0 access token from go-oauthv2. Format: Bearer {access_token}

package main

import (
	"context"
	"log"

	"github.com/overiss/vectovm-api/internal/app"
	"github.com/overiss/vectovm-api/internal/config"
	"github.com/overiss/vectovm-api/pkg/utils"
	"golang.org/x/sync/errgroup"
)

//go:generate swag init -g main.go -o ../../api/docs --parseInternal --parseDependency

func main() {
	cfg := config.Init()
	erg, ctx := errgroup.WithContext(context.Background())

	erg.Go(func() error {
		return utils.Listen(ctx)
	})

	log.Printf("config initialized")

	application := app.Init(ctx, cfg)
	log.Printf("service initialized successfully")
	go application.Start(ctx)

	if err := erg.Wait(); err != nil {
		log.Printf("stopping application cause: %v", err)
	}

	application.Stop(ctx)
	log.Printf("application stopped gracefully")
}
