package server

import serverHttp "github.com/overiss/vectovm-api/internal/server/http"

type Container struct {
	http *serverHttp.Http
}

func NewContainer(http *serverHttp.Http) *Container {
	return &Container{http: http}
}

func (c *Container) Http() *serverHttp.Http {
	return c.http
}
