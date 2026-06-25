package utils

import (
	"context"
	"os/signal"
	"syscall"
)

func Listen(ctx context.Context) error {
	sigCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-sigCtx.Done()
	if sigCtx.Err() == context.Canceled && ctx.Err() == nil {
		return nil
	}
	if sigCtx.Err() == context.Canceled {
		return nil
	}
	return sigCtx.Err()
}
