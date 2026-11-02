package http

import (
	"bytes"
	"encoding/base64"
	"gohole/internal/controller/dns"
	"io"
	"net"
	"net/http"

	gdns "codeberg.org/miekg/dns"
	"github.com/go-chi/chi/v5"
)

type DoHRouter struct {
	dnsHandler *dns.Handler
}

func NewDoHRouter(dh *dns.Handler) *DoHRouter {
	return &DoHRouter{dnsHandler: dh}
}

func (dr *DoHRouter) Register(r chi.Router) {
	r.Post("/dns-query", errorHandler(dr.handleDoH))
	r.Get("/dns-query", errorHandler(dr.handleDoH))
}

func (dr *DoHRouter) handleDoH(w http.ResponseWriter, r *http.Request) error {
	var body []byte
	var err error

	if r.Method == http.MethodPost {
		body, err = io.ReadAll(r.Body)
		if err != nil {
			return newHTTPErr(http.StatusBadRequest, "Failed to read request body")
		}
	} else if r.Method == http.MethodGet {
		dnsParam := r.URL.Query().Get("dns")
		if dnsParam == "" {
			return newHTTPErr(http.StatusBadRequest, "Missing 'dns' query parameter")
		}
		body, err = base64.RawURLEncoding.DecodeString(dnsParam)
		if err != nil {
			body, err = base64.URLEncoding.DecodeString(dnsParam)
			if err != nil {
				return newHTTPErr(http.StatusBadRequest, "Failed to decode 'dns' query parameter")
			}
		}
	} else {
		return newHTTPErr(http.StatusMethodNotAllowed, "Method not allowed")
	}

	reqMsg := new(gdns.Msg)
	reqMsg.Data = body
	if err := reqMsg.Unpack(); err != nil {
		return newHTTPErr(http.StatusBadRequest, "Failed to unpack DNS message")
	}

	clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		clientIP = r.RemoteAddr
	}

	remoteAddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(clientIP, "53"))
	if err != nil {
		remoteAddr = &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 53}
	}

	dohWriter := &dohResponseWriter{
		remoteAddr: remoteAddr,
	}

	dr.dnsHandler.HandleRequestWithMiddlewares(r.Context(), dohWriter, reqMsg)

	respBytes := dohWriter.buf.Bytes()
	if len(respBytes) > 2 {
		respBytes = respBytes[2:]
	}

	w.Header().Set("Content-Type", "application/dns-message")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(respBytes)
	return nil
}

// dohResponseWriter is a custom ResponseWriter that captures the DNS response
// in a buffer.
type dohResponseWriter struct {
	remoteAddr net.Addr
	buf        bytes.Buffer
}

func (w *dohResponseWriter) LocalAddr() net.Addr {
	return &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 53}
}
func (w *dohResponseWriter) RemoteAddr() net.Addr {
	return w.remoteAddr
}
func (w *dohResponseWriter) Conn() net.Conn {
	return nil
}
func (w *dohResponseWriter) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}
func (w *dohResponseWriter) Close() error {
	return nil
}
func (w *dohResponseWriter) Session() *gdns.Session {
	return nil
}
func (w *dohResponseWriter) Hijack() {}
