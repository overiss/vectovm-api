package behavior

import "context"

type Starter interface {
	Start(ctx context.Context)
	Stop(ctx context.Context)
}

type Readiness interface {
	Name() string
	IsReady() bool
}
