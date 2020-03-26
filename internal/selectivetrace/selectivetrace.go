// Package selectivetrace wraps the opentracing API to trace selectively. Whether or not tracing
// occurs depends on the presence of a context item.
//
// If the context item is true, then we use opentracing.GlobalTracer. Otherwise,
// an instance of opentracing.NoopTracer is used.
//
// Motivation: this package exists for performance reasons. In large instances of Sourcegraph, we
// have seen Jaeger tracing overwhelm the network due to the "fan-out" behavior of our search
// backend. This occurs even if subsampling is enabled in Jaeger, indicating the bottleneck exists
// between the Jaeger client (i.e., Sourcegraph) and the Jaeger agent.
package selectivetrace

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

func GlobalTracer(ctx context.Context) ot.Tracer {
	if FromContext(ctx) {
		return ot.GlobalTracer()
	}
	return ot.NoopTracer{}
}

func StartSpan(operationName string, opts ...ot.StartSpanOption) ot.Span {
	if FromContext(ctx) {
		return ot.StartSpan(operationName, opts...)
	}
	return ot.NoopTracer{}.StartSpan(operationName, opts...)
}

func StartSpanFromContext(ctx context.Context, operationName string, opts ...ot.StartSpanOption) (ot.Span, context.Context) {
	if FromContext(ctx) {
		return ot.StartSpanFromContext(ctx, operationName, opts...)
	}
	return ot.StartSpanFromContextWithTracer(ctx, ot.NoopTracer{}, operationName, opts...)
}
