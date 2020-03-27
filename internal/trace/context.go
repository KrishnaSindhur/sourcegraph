package trace

import (
	"context"
	"net/http"
	"strconv"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
	ot "github.com/opentracing/opentracing-go"
)

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
	return GetTracerNonGlobal(ctx, ot.GlobalTracer())
}

func GetTracerNonGlobal(ctx context.Context, tracer ot.Tracer) ot.Tracer {
	// TODO: incorporate config value, via init func and conf.Watch
	if FromContext(ctx) {
		return tracer
	}
	return ot.NoopTracer{}

}

// StartSpanFromContext conditionally starts a span either with the global tracer or the NoopTracer,
// depending on whether the context item is set and if selective tracing is enabled in the site
// configuration.
func StartSpanFromContext(ctx context.Context, operationName string, opts ...ot.StartSpanOption) (ot.Span, context.Context) {
	return ot.StartSpanFromContextWithTracer(ctx, GetTracer(ctx), operationName, opts...)
}

func Middleware(h http.Handler, opts ...nethttp.MWOption) http.Handler {
	return MiddlewareWithTracer(ot.GlobalTracer(), h)
}

func MiddlewareWithTracer(tr opentracing.Tracer, h http.Handler, opts ...nethttp.MWOption) http.Handler {
	// TODO: incorporate config value, via init func and conf.Watch
	allOpts := append([]nethttp.MWOption{
		nethttp.MWSpanFilter(func(r *http.Request) bool { return FromContext(r.Context()) }),
	}, opts...)
	m := nethttp.Middleware(tr, h, allOpts...)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if shouldTrace, _ := strconv.ParseBool(r.URL.Query().Get("trace")); shouldTrace {
			m.ServeHTTP(w, r.WithContext(WithTracing(r.Context(), true)))
		}
		m.ServeHTTP(w, r)
	})
}
