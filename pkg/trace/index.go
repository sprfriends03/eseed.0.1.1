package trace

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/nhnghia272/gopkg"
)

type contextKey string

const traceID = contextKey("traceID")

type E struct {
	K string
	V any
}

func (e E) String() string {
	val, _ := json.Marshal(e.V)
	return fmt.Sprintf("%s: %s", e.K, string(val))
}

var trace = gopkg.NewCacheShard[[]E](64)

func New(ctx context.Context) context.Context {
	return context.WithValue(ctx, traceID, uuid.NewString())
}

func WithValue(ctx context.Context, values ...E) context.Context {
	if id := ctx.Value(traceID); id != nil && len(values) > 0 {
		trace.Set(id.(string), append(Value(ctx), values...), -1)
	}
	return ctx
}

func Value(ctx context.Context) []E {
	if id := ctx.Value(traceID); id != nil {
		items, _ := trace.Get(id.(string))
		return items
	}
	return make([]E, 0)
}

func Clear(ctx context.Context) {
	if id := ctx.Value(traceID); id != nil {
		trace.Delete(id.(string))
	}
}

func ID(ctx context.Context) string {
	if id := ctx.Value(traceID); id != nil {
		return id.(string)
	}
	return ""
}
