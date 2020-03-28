package trace

import (
	"context"
	"net/http"
	"strconv"

	"github.com/inconshreveable/log15"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
	ot "github.com/opentracing/opentracing-go"
)

var (
	AlwaysTrace = false
	DebugLog    = true
)

func shouldTrace(ctx context.Context) bool {
	if AlwaysTrace {
		return true
	}
	return FromContext(ctx)
}

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
	if shouldTrace(ctx) {
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
	m := nethttp.Middleware(tr, h, append([]nethttp.MWOption{
		nethttp.MWSpanFilter(func(r *http.Request) bool {
			if DebugLog {
				log15.Info("trace: MiddlewareWithTracer", "url", r.URL.String(), "shouldTrace", shouldTrace(r.Context()))
			}
			return shouldTrace(r.Context())
		}),
	}, opts...)...)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if traceHeaderIsTrue, _ := strconv.ParseBool(r.Header.Get(traceHeader)); traceHeaderIsTrue {
			m.ServeHTTP(w, r.WithContext(WithTracing(r.Context(), true)))
			return
		}
		m.ServeHTTP(w, r)
	})
}

const traceHeader = "X-Sourcegraph-Trace"

// RequestWithContextHeader modifies the original header to set the HTTP header "X-Sourcegraph-Trace".
// The input request (which is modified) is returned.
func RequestWithContextHeader(ctx context.Context, r *http.Request) *http.Request {
	r.Header.Set(traceHeader, strconv.FormatBool(FromContext(ctx)))
	return r
}
