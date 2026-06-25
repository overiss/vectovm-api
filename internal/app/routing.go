package app

import (
	"context"

	hanlderHttp "github.com/overiss/vectovm-api/internal/server/http/handler"
)

func (a *Application) expose(ctx context.Context) {
	if a.handlerContainer == nil || a.netContainer == nil {
		return
	}
	hanlderHttp.RegisterRoutes(a.netContainer.Http().Router(), a.handlerContainer)
}
