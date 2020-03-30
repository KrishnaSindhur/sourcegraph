// Package ot wraps github.com/opentracing/opentracing-go and
// github.com./opentracing-contrib/go-stdlib with selective tracing behavior that is toggled on and
// off with the presence of a context item (uses context.Context). This context item is propagated
// across API boundaries through a HTTP header (X-Sourcegraph-Trace).
package ot

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
)

type key int

const (
	shouldTraceKey key = iota
)

type SamplingStrategy string

const (
	SampleNone      SamplingStrategy = "none"
	SampleSelective                  = "selective"
	SampleAll                        = "comprehensive"
)

var (
	Sampling SamplingStrategy = "none"
	DebugLog                  = true
)

func ShouldTrace(ctx context.Context) bool {
	v, ok := ctx.Value(shouldTraceKey).(bool)
	if !ok {
		return false
	}
	return v
}

// WithShouldTrace sets the tracing context item, which will enable traces on operations that use the context.
func WithShouldTrace(ctx context.Context, shouldTrace bool) context.Context {
	return context.WithValue(ctx, shouldTraceKey, shouldTrace)
}

func GetTracer(ctx context.Context) opentracing.Tracer {
	return getTracer(ctx, opentracing.GlobalTracer())
}

func getTracer(ctx context.Context, tracer opentracing.Tracer) opentracing.Tracer {
	if !ShouldTrace(ctx) {
		return opentracing.NoopTracer{}
	}
	if tracer != nil {
		return tracer
	}
	return opentracing.GlobalTracer()
}

// StartSpanFromContext conditionally starts a span either with the global tracer or the NoopTracer,
// depending on whether the context item is set and if selective tracing is enabled in the site
// configuration.
func StartSpanFromContext(ctx context.Context, operationName string, opts ...opentracing.StartSpanOption) (opentracing.Span, context.Context) {
	return StartSpanFromContextWithTracer(ctx, opentracing.GlobalTracer(), operationName, opts...)
}

func StartSpanFromContextWithTracer(ctx context.Context, tracer opentracing.Tracer, operationName string, opts ...opentracing.StartSpanOption) (opentracing.Span, context.Context) {
	return opentracing.StartSpanFromContextWithTracer(ctx, getTracer(ctx, tracer), operationName, opts...)
}

func Middleware(h http.Handler, opts ...nethttp.MWOption) http.Handler {
	return MiddlewareWithTracer(opentracing.GlobalTracer(), h)
}

func MiddlewareWithTracer(tr opentracing.Tracer, h http.Handler, opts ...nethttp.MWOption) http.Handler {
	m2 := nethttp.Middleware(tr, h, append([]nethttp.MWOption{
		nethttp.MWSpanFilter(func(r *http.Request) bool {
			return ShouldTrace(r.Context())
		}),
	}, opts...)...)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("# Sampling %v", Sampling)
		switch Sampling {
		case "selective":
			traceHeaderIsTrue, _ := strconv.ParseBool(r.Header.Get(traceHeader))
			m2.ServeHTTP(w, r.WithContext(WithShouldTrace(r.Context(), traceHeaderIsTrue)))
			return
		case "comprehensive":
			m2.ServeHTTP(w, r.WithContext(WithShouldTrace(r.Context(), true)))
			return
		default:
			m2.ServeHTTP(w, r.WithContext(WithShouldTrace(r.Context(), false)))
			return
		}
	})
}

const traceHeader = "X-Sourcegraph-Trace"

type Transport struct {
	http.RoundTripper
}

func (r *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set(traceHeader, strconv.FormatBool(ShouldTrace(req.Context())))
	t := nethttp.Transport{RoundTripper: r.RoundTripper}
	return t.RoundTrip(req)
}
