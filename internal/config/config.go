package config

import (
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

var Version = "v1.0.0"

type (
	Application struct {
		Http     *HttpServer `envconfig:"HTTP" required:"true"`
		Postgres Postgres    `envconfig:"POSTGRES" required:"true"`
		OAuth    OAuth       `envconfig:"OAUTH" required:"true"`
		Mapi     Mapi        `envconfig:"MAPI" required:"true"`
		Vault    Vault       `envconfig:"VAULT" required:"true"`

		Version string
		IsDebug bool `envconfig:"IS_DEBUG"`
	}

	HttpServer struct {
		Name         string        `envconfig:"NAME" required:"true"`
		Port         string        `envconfig:"PORT" required:"true"`
		ReadTimeout  time.Duration `envconfig:"READ_TIMEOUT" default:"15s"`
		WriteTimeout time.Duration `envconfig:"WRITE_TIMEOUT" default:"30s"`
		IdleTimeout  time.Duration `envconfig:"IDLE_TIMEOUT" default:"10s"`
	}

	Postgres struct {
		DSN             string        `envconfig:"DSN" required:"true"`
		MaxConns        int32         `envconfig:"MAX_CONNS" default:"10"`
		MinConns        int32         `envconfig:"MIN_CONNS" default:"2"`
		MaxConnLifetime time.Duration `envconfig:"MAX_CONN_LIFETIME" default:"1h"`
		MaxConnIdleTime time.Duration `envconfig:"MAX_CONN_IDLE_TIME" default:"30m"`
	}

	OAuth struct {
		Issuer       string `envconfig:"ISSUER" required:"true"`
		ClientID     string `envconfig:"CLIENT_ID" required:"true"`
		ClientSecret string `envconfig:"CLIENT_SECRET" required:"true"`
		RedirectURI  string `envconfig:"REDIRECT_URI" required:"true"`
	}

	Mapi struct {
		BaseURL     string `envconfig:"BASE_URL" required:"true"`
		BearerToken string `envconfig:"BEARER_TOKEN" required:"true"`
		CACertFile  string `envconfig:"CA_CERT_FILE"`
	}

	Vault struct {
		Address            string `envconfig:"ADDRESS" required:"true"`
		Token              string `envconfig:"TOKEN" required:"true"`
		CredentialsKeyPath string `envconfig:"CREDENTIALS_KEY_PATH" default:"secret/data/vectovm-api/credentials"`
	}
)

func Init() *Application {
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: error loading .env file")
	}

	app := new(Application)
	if err := envconfig.Process("", app); err != nil {
		log.Fatalf("cannot process config for vectovm-api: %v", err)
	}

	app.Version = Version
	return app
}
