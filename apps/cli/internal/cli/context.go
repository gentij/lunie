package cli

import (
	"context"

	"github.com/gentij/lunie/apps/cli/internal/api"
	"github.com/gentij/lunie/apps/cli/internal/config"
)

type Context struct {
	Config  config.Config
	Client  *api.Client
	Output  string
	Quiet   bool
	NoColor bool
}

type ctxKey struct{}

func WithContext(ctx context.Context, value *Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, value)
}

func GetContext(ctx context.Context) *Context {
	value := ctx.Value(ctxKey{})
	if value == nil {
		return nil
	}

	if typed, ok := value.(*Context); ok {
		return typed
	}

	return nil
}

func IsJSON(ctx *Context) bool {
	return ctx != nil && ctx.Output == "json"
}
