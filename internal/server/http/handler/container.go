package hanlderHttp

import (
	"github.com/overiss/vectovm-api/internal/auth"
)

type Container struct {
	Health   *HealthHandler
	Auth     *AuthHandler
	User     *UserHandler
	Datanode *DatanodeHandler
	VM       *VMHandler
	Verifier *auth.Verifier
}

func NewContainer(
	health *HealthHandler,
	authHandler *AuthHandler,
	user *UserHandler,
	datanode *DatanodeHandler,
	vm *VMHandler,
	verifier *auth.Verifier,
) *Container {
	return &Container{
		Health:   health,
		Auth:     authHandler,
		User:     user,
		Datanode: datanode,
		VM:       vm,
		Verifier: verifier,
	}
}
