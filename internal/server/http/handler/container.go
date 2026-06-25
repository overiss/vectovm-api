package hanlderHttp

import (
	"github.com/overiss/vectovm-api/internal/auth"
)

type Container struct {
	Auth     *AuthHandler
	User     *UserHandler
	Datanode *DatanodeHandler
	VM       *VMHandler
	Verifier *auth.Verifier
}

func NewContainer(
	authHandler *AuthHandler,
	user *UserHandler,
	datanode *DatanodeHandler,
	vm *VMHandler,
	verifier *auth.Verifier,
) *Container {
	return &Container{
		Auth:     authHandler,
		User:     user,
		Datanode: datanode,
		VM:       vm,
		Verifier: verifier,
	}
}
