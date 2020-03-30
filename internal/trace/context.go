package trace

import (
	"context"
	"net/http"
	"strconv"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	ot "github.com/opentracing/opentracing-go"
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

func fromContext(ctx context.Context) bool {
	v, ok := ctx.Value(shouldTraceKey).(bool)
	if !ok {
		return false
	}
	return v
}

// withTracing sets the tracing context item, which will enable traces on operations that use the context.
func withTracing(ctx context.Context, shouldTrace bool) context.Context {
	return context.WithValue(ctx, shouldTraceKey, shouldTrace)
}

func GetTracer(ctx context.Context) ot.Tracer {
	return getTracer(ctx, ot.GlobalTracer())
}

func getTracer(ctx context.Context, tracer ot.Tracer) ot.Tracer {
	if !fromContext(ctx) {
		return ot.NoopTracer{}
	}
	if tracer != nil {
		return tracer
	}
	return ot.GlobalTracer()
}

// StartSpanFromContext conditionally starts a span either with the global tracer or the NoopTracer,
// depending on whether the context item is set and if selective tracing is enabled in the site
// configuration.
func StartSpanFromContext(ctx context.Context, operationName string, opts ...ot.StartSpanOption) (ot.Span, context.Context) {
	return StartSpanFromContextWithTracer(ctx, ot.GlobalTracer(), operationName, opts...)
}

func StartSpanFromContextWithTracer(ctx context.Context, tracer ot.Tracer, operationName string, opts ...ot.StartSpanOption) (ot.Span, context.Context) {
	return ot.StartSpanFromContextWithTracer(ctx, getTracer(ctx, tracer), operationName, opts...)
}

func Middleware(h http.Handler, opts ...nethttp.MWOption) http.Handler {
	return MiddlewareWithTracer(ot.GlobalTracer(), h)
}

func MiddlewareWithTracer(tr ot.Tracer, h http.Handler, opts ...nethttp.MWOption) http.Handler {
	m2 := nethttp.Middleware(tr, h, append([]nethttp.MWOption{
		nethttp.MWSpanFilter(func(r *http.Request) bool {
			return fromContext(r.Context())
		}),
	}, opts...)...)
	// // logging
	// m2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	if DebugLog {
	// 		log15.Info("trace: MiddlewareWithTracer", "url", r.URL.String(), "shouldTrace", shouldTrace(r.Context()))
	// 	}
	// 	m.ServeHTTP(w, r)
	// })
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch Sampling {
		case "selective":
			traceHeaderIsTrue, _ := strconv.ParseBool(r.Header.Get(traceHeader))
			m2.ServeHTTP(w, r.WithContext(withTracing(r.Context(), traceHeaderIsTrue)))
			return
		case "comprehensive":
			m2.ServeHTTP(w, r.WithContext(withTracing(r.Context(), true)))
			return
		default:
			m2.ServeHTTP(w, r.WithContext(withTracing(r.Context(), false)))
			return
		}
	})
}

const traceHeader = "X-Sourcegraph-Trace"

type Transport struct {
	http.RoundTripper
}

func (r *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set(traceHeader, strconv.FormatBool(fromContext(req.Context())))
	t := nethttp.Transport{RoundTripper: r.RoundTripper}
	return t.RoundTrip(req)
}
