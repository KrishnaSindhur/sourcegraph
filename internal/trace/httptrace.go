package trace

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/inconshreveable/log15"

	raven "github.com/getsentry/raven-go"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

type key int

const (
	routeNameKey key = iota
	userKey
	requestErrorCauseKey
	graphQLRequestNameKey
	originKey
	contextKey
)

// trackOrigin specifies a URL value. When an incoming request has the request header "Origin" set
// and the header value equals the `trackOrigin` value then the `requestDuration` metric (and other metrics downstream)
// gets labeled with this value for the "origin" label  (otherwise the metric is labeled with "unknown").
// The tracked value can be changed with the METRICS_TRACK_ORIGIN environmental variable.
var trackOrigin = "https://gitlab.com"

var metricLabels = []string{"route", "method", "code", "repo", "origin"}
var requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "src",
	Subsystem: "http",
	Name:      "request_duration_seconds",
	Help:      "The HTTP request latencies in seconds.",
	Buckets:   UserLatencyBuckets,
}, metricLabels)

var requestHeartbeat = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "http",
	Name:      "requests_last_timestamp_unixtime",
	Help:      "Last time a request finished for a http endpoint.",
}, metricLabels)

func init() {
	if err := raven.SetDSN(os.Getenv("SENTRY_DSN_BACKEND")); err != nil {
		log15.Error("sentry.dsn", "error", err)
	}

	if origin := os.Getenv("METRICS_TRACK_ORIGIN"); origin != "" {
		trackOrigin = origin
	}

	raven.SetRelease(version.Version())
	raven.SetTagsContext(map[string]string{
		"service": filepath.Base(os.Args[0]),
	})

	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(requestHeartbeat)

	go func() {
		conf.Watch(func() {
			if conf.Get().Log == nil {
				return
			}

			if conf.Get().Log.Sentry == nil {
				return
			}

			// An empty dsn value is ignored: not an error.
			if err := raven.SetDSN(conf.Get().Log.Sentry.Dsn); err != nil {
				log15.Error("sentry.dsn", "error", err)
			}
		})
	}()
}

// GraphQLRequestName returns the GraphQL request name for a request context. For example,
// a request to /.api/graphql?Foobar would have the name `Foobar`. If the request had no
// name, or the context is not a GraphQL request, "unknown" is returned.
func GraphQLRequestName(ctx context.Context) string {
	v := ctx.Value(graphQLRequestNameKey)
	if v == nil {
		return "unknown"
	}
	return v.(string)
}

// WithGraphQLRequestName sets the GraphQL request name in the context.
func WithGraphQLRequestName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, graphQLRequestNameKey, name)
}

// RequestOrigin returns the request origin (the value of the request header "Origin") for a request context.
// If the request didn't have this header set "unknown" is returned.
func RequestOrigin(ctx context.Context) string {
	v := ctx.Value(originKey)
	if v == nil {
		return "unknown"
	}
	return v.(string)
}

// WithRequestOrigin sets the request origin in the context.
func WithRequestOrigin(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, originKey, name)
}

func TraceRoute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if p, ok := r.Context().Value(routeNameKey).(*string); ok {
			if routeName := mux.CurrentRoute(r).GetName(); routeName != "" {
				*p = routeName
			}
		}
		next.ServeHTTP(rw, r)
	})
}

func TraceUser(ctx context.Context, userID int32) {
	if p, ok := ctx.Value(userKey).(*int32); ok {
		*p = userID
	}
}

// SetRequestErrorCause will set the error for the request to err. This is
// used in the reporting layer to inspect the error for richer reporting to
// Sentry.
func SetRequestErrorCause(ctx context.Context, err error) {
	if p, ok := ctx.Value(requestErrorCauseKey).(*error); ok {
		*p = err
	}
}

// SetRouteName manually sets the name for the route. This should only be used
// for non-mux routed routes (ie middlewares).
func SetRouteName(r *http.Request, routeName string) {
	if p, ok := r.Context().Value(routeNameKey).(*string); ok {
		*p = routeName
	}
}

type httpErr struct {
	status int
	method string
	path   string
}

func (e *httpErr) Error() string {
	return fmt.Sprintf("HTTP status %d, %s %s", e.status, e.method, e.path)
}
