package dns

import (
	"context"
	"log/slog"
	"net"
	"slices"
	"sync"
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

func (r *ReqCtx) Free() {
	r.Context = nil
	r.Logger = nil
	r.Trace = ""
	r.Start = time.Time{}
	r.End = time.Time{}
	r.Name = ""
	r.Host = ""
	r.Allowed = false
	r.Cached = false
	r.Custom = false
	r.Error = nil
}

var ctxPool = sync.Pool{
	New: func() any {
		return &ReqCtx{}
	},
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

		rc := ctxPool.Get().(*ReqCtx)
		rc.Context = ctx
		rc.Host = host
		rc.Trace = trace
		rc.Logger = l

		f := handler
		for _, m := range slices.Backward(middlewares) {
			f = m(f)
		}

		f(rc, w, msg)

		rc.Free()
		ctxPool.Put(rc)
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
