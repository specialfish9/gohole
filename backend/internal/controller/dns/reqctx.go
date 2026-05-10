package dns

import (
	"context"
	"log/slog"
	"net"
	"slices"
	"time"

	"codeberg.org/miekg/dns"
)

const reqCtxKey = "reqCtx"

type ReqCtx struct {
	Context context.Context
	Logger  *slog.Logger
	Trace   string
	Start   time.Time
	End     time.Time
	Name    string
	Host    string
	Allowed bool
	Cached  bool
	Custom  bool
	Error   error
}

// applyMiddlewares initializes the request context and builds the middleware chain
func applyMiddlewares(handler handlerFunc, middlewares ...middleware) dns.HandlerFunc {
	return func(ctx context.Context, w dns.ResponseWriter, msg *dns.Msg) {
		trace := freshTraceID()
		l := slog.With("trace", trace)

		host, _, err := net.SplitHostPort(w.RemoteAddr().String())
		if err != nil {
			l.Error("Failed to parse client address", "error", err.Error(), "remoteAddr", w.RemoteAddr().String())
			host = w.RemoteAddr().String()
		}

		rc := &ReqCtx{
			Context: ctx,
			Logger:  l,
			Host:    host,
			Trace:   trace,
		}

		f := handler
		for _, m := range slices.Backward(middlewares) {
			f = m(f)
		}

		f(rc, w, msg)
	}
}

func recoverMiddleware(next handlerFunc) handlerFunc {
	return func(rc *ReqCtx, w dns.ResponseWriter, r *dns.Msg) {
		defer func() {
			if err := recover(); err != nil {
				rc.Logger.Error("PANIC!", "message", err)
			}
		}()

		next(rc, w, r)
	}
}

func logMiddleware(kv ...any) middleware {
	return func(next handlerFunc) handlerFunc {
		return func(rc *ReqCtx, w dns.ResponseWriter, r *dns.Msg) {
			rc.Logger = rc.Logger.With(kv...)
			next(rc, w, r)
			if rc.Error != nil {
				rc.Logger.Error(
					rc.Error.Error(),
					"name", rc.Name,
					"trace", rc.Trace,
					"host", rc.Host,
				)
			} else {
				var mex string
				if rc.Allowed {
					mex = "PASS"
				} else {
					mex = "SMASH"
				}

				rc.Logger.Info(
					mex,
					"name", rc.Name,
					"trace", rc.Trace,
					"timeMicro", rc.End.Sub(rc.Start).Microseconds(),
					"host", rc.Host,
					"cache", rc.Cached,
					"customDomain", rc.Custom,
				)
			}
		}
	}
}

func timeMiddleware(next handlerFunc) handlerFunc {
	return func(rc *ReqCtx, w dns.ResponseWriter, r *dns.Msg) {
		rc.Start = time.Now()
		next(rc, w, r)
		rc.End = time.Now()
	}
}
