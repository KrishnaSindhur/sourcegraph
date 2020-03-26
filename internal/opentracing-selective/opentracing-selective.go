// Package opentracing wraps github.com/opentracing/opentracing-go to trace selectively, depending
// on the presence of a context item. If the context item is not set or contains the value, false,
// tracing functions that reference the global tracer (set by
// github.com/opentracing/opentracing-go.SetTracer) use an instance of
// github.com/opentracing/opentracing-go.NoopTracer.
//
// This package exists for performance reasons. In large instances of Sourcegraph, distributed
// tracing may overwhelm the network due to the "fan-out" behavior of our search backend. This
// occurs even if subsampling is enabled in the trace collector (because this does not affect
// traffic between the tracing client and tracing system (e.g., between the Sourcegraph process and
// the jaeger-agent).
package opentracing

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

func StartSpanFromContext(ctx context.Context, operationName string, opts ...ot.StartSpanOption) (ot.Span, context.Context) {
	if FromContext(ctx) {
		return ot.StartSpanFromContext(ctx, operationName, opts...)
	}
	return ot.StartSpanFromContextWithTracer(ctx, ot.NoopTracer{}, operationName, opts...)
}
