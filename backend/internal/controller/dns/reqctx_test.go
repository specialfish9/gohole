package dns

import (
	"context"
	"testing"
	"time"

	gdns "codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnstest"
)

func TestApplyMiddlewares(t *testing.T) {
	t.Run("populates ReqCtx from context and writer", func(t *testing.T) {
		// Capture a snapshot of the ReqCtx before the pool recycles it.
		var captured ReqCtx
		handler := func(rc *ReqCtx, w gdns.ResponseWriter, r *gdns.Msg) {
			captured = *rc
		}

		ctx := context.Background()
		// The default RemoteAddr is 198.51.100.1:40212
		w := &dnstest.ResponseWriter{}
		msg := gdns.NewMsg("example.com", gdns.TypeA)

		fn := applyMiddlewares(handler)
		fn(ctx, w, msg)

		if captured.Host != "198.51.100.1" {
			t.Errorf("expected Host %q, got %q", "198.51.100.1", captured.Host)
		}
		if captured.Trace == "" {
			t.Error("expected Trace to be set, got empty string")
		}
		if captured.Context != ctx {
			t.Error("expected Context to be the one passed to the HandlerFunc")
		}
		if captured.Logger == nil {
			t.Error("expected Logger to be set")
		}
	})

	t.Run("applies middlewares outermost-first (m1 wraps m2 wraps handler)", func(t *testing.T) {
		var callOrder []string

		makeMiddleware := func(name string) middleware {
			return func(next handlerFunc) handlerFunc {
				return func(rc *ReqCtx, w gdns.ResponseWriter, r *gdns.Msg) {
					callOrder = append(callOrder, name+":enter")
					next(rc, w, r)
					callOrder = append(callOrder, name+":exit")
				}
			}
		}

		handler := func(rc *ReqCtx, w gdns.ResponseWriter, r *gdns.Msg) {
			callOrder = append(callOrder, "handler")
		}

		fn := applyMiddlewares(handler, makeMiddleware("m1"), makeMiddleware("m2"))
		fn(context.Background(), &dnstest.ResponseWriter{}, gdns.NewMsg("example.com", gdns.TypeA))

		// m1 is the outermost wrapper, so it enters first and exits last.
		want := []string{"m1:enter", "m2:enter", "handler", "m2:exit", "m1:exit"}
		if len(callOrder) != len(want) {
			t.Fatalf("call order: want %v, got %v", want, callOrder)
		}
		for i, v := range want {
			if callOrder[i] != v {
				t.Errorf("step %d: want %q, got %q", i, v, callOrder[i])
			}
		}
	})
}

func TestTimeMiddleware(t *testing.T) {
	var startWhenHandlerRan time.Time
	var endWhenHandlerRan time.Time

	handler := func(rc *ReqCtx, w gdns.ResponseWriter, r *gdns.Msg) {
		// Capture both timestamps while the handler is still executing.
		startWhenHandlerRan = rc.Start
		endWhenHandlerRan = rc.End
	}

	rc := &ReqCtx{}
	timeMiddleware(handler)(rc, nil, nil)

	// Start must be populated before the handler is invoked.
	if startWhenHandlerRan.IsZero() {
		t.Error("expected rc.Start to be set before handler ran")
	}
	// End must not be set yet while the handler is running.
	if !endWhenHandlerRan.IsZero() {
		t.Error("expected rc.End to still be zero while handler is running")
	}
	// After the call both fields must be non-zero.
	if rc.Start.IsZero() {
		t.Error("expected rc.Start to be non-zero after call")
	}
	if rc.End.IsZero() {
		t.Error("expected rc.End to be non-zero after call")
	}
	// End must not precede Start.
	if rc.End.Before(rc.Start) {
		t.Errorf("expected End (%v) to not be before Start (%v)", rc.End, rc.Start)
	}
}
