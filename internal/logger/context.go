package logger

import (
	"context"
	"log"
)

type contextKeyType string

var (
	contextKey = contextKeyType("logger")
)

func FromContext(ctx context.Context) Logger {
	if ctx != nil {
		if v := ctx.Value(contextKey); v != nil {
			return v.(Logger)
		}
	}

	log.Fatal("No logger found in context")
	return nil
}

func OnContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, contextKey, logger)
}
