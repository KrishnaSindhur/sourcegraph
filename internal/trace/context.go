package trace

import (
	"context"

	ot "github.com/opentracing/opentracing-go"
)

type key int

const contextKey key = iota

func FromContext(ctx context.Context) bool {
	v, ok := ctx.Value(contextKey).(bool)
	if !ok {
		return false
	}
	return v
}

func WithTracing(ctx context.Context, shouldTrace bool) context.Context {
	return context.WithValue(ctx, contextKey, shouldTrace)
}

func GetTracer(ctx context.Context) ot.Tracer {
	// TODO: incorporate config value, via init func and conf.Watch
	if FromContext(ctx) {
		return ot.GlobalTracer()
	}
	return ot.NoopTracer{}
}

// StartSpanFromContext conditionally starts a span either with the global tracer or the NoopTracer,
// depending on whether the context item is set and if selective tracing is enabled in the site
// configuration.
func StartSpanFromContext(ctx context.Context, operationName string, opts ...ot.StartSpanOption) (ot.Span, context.Context) {
	return ot.StartSpanFromContextWithTracer(ctx, ot.GetTracer(ctx), operationName, opts...)
}

// TODO: Middleware

// >>>>>>>>> Replace all instances of selectivetracing.StartSpanFromContext with StartSpanFromContext
