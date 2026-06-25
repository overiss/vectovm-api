package service

import (
	"context"
	"fmt"

	"github.com/overiss/vectovm-api/internal/auth"
	mapiclient "github.com/overiss/vectovm-api/internal/client/mapi"
	oauthclient "github.com/overiss/vectovm-api/internal/client/oauth"
	"github.com/overiss/vectovm-api/internal/config"
	"github.com/overiss/vectovm-api/internal/crypto"
	"github.com/overiss/vectovm-api/internal/repository"
	authservice "github.com/overiss/vectovm-api/internal/service/auth"
	datanodeservice "github.com/overiss/vectovm-api/internal/service/datanode"
	userservice "github.com/overiss/vectovm-api/internal/service/user"
	vmservice "github.com/overiss/vectovm-api/internal/service/vm"
	"github.com/overiss/vectovm-api/internal/storage/postgres"
	vaultstore "github.com/overiss/vectovm-api/internal/vault"
)

type Container struct {
	Auth     *authservice.Service
	User     *userservice.Service
	Datanode *datanodeservice.Service
	VM       *vmservice.Service
	Verifier *auth.Verifier

	db *postgres.DB
}

func NewContainer(ctx context.Context, cfg *config.Application) (*Container, error) {
	db, err := postgres.New(ctx, cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("init postgres: %w", err)
	}

	keyStore, err := vaultstore.NewKeyStore(cfg.Vault.Address, cfg.Vault.Token, cfg.Vault.CredentialsKeyPath)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("init vault keystore: %w", err)
	}

	masterKey, err := keyStore.CredentialsMasterKey(ctx)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("load credentials master key: %w", err)
	}

	envelope, err := crypto.NewEnvelope(masterKey)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("init envelope encryption: %w", err)
	}

	oauth := oauthclient.NewClient(
		cfg.OAuth.Issuer,
		cfg.OAuth.ClientID,
		cfg.OAuth.ClientSecret,
		cfg.OAuth.RedirectURI,
	)

	mapi, err := mapiclient.NewClient(cfg.Mapi)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("init mapi client: %w", err)
	}

	userRepo := repository.NewUserRepository(db)
	datanodeRepo := repository.NewDatanodeRepository(db)
	vmRepo := repository.NewVMRepository(db)
	credentials := crypto.NewCredentialService(envelope, userRepo)

	return &Container{
		Auth:     authservice.NewService(oauth, userRepo, credentials),
		User:     userservice.NewService(userRepo),
		Datanode: datanodeservice.NewService(mapi, userRepo, datanodeRepo),
		VM:       vmservice.NewService(vmRepo, datanodeRepo, userRepo, credentials),
		Verifier: auth.NewVerifier(oauth.JWKSURL()),
		db:       db,
	}, nil
}

func (c *Container) Close() {
	if c.db != nil {
		c.db.Close()
	}
}

func (c *Container) DB() *postgres.DB {
	return c.db
}
