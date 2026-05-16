package dns_test

import (
	"bytes"
	"fmt"
	"net"

	gdns "codeberg.org/miekg/dns"
)

// fakeWriter captures all bytes written by the handler so we can parse the DNS response.
// It implements codeberg.org/miekg/dns.ResponseWriter.
type fakeWriter struct {
	buf bytes.Buffer
}

func (w *fakeWriter) LocalAddr() net.Addr {
	return &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 53}
}
func (w *fakeWriter) RemoteAddr() net.Addr {
	return &net.UDPAddr{IP: net.ParseIP("198.51.100.1"), Port: 40212}
}
func (w *fakeWriter) Conn() net.Conn              { return nil }
func (w *fakeWriter) Write(b []byte) (int, error) { return w.buf.Write(b) }
func (w *fakeWriter) Close() error                { return nil }
func (w *fakeWriter) Session() *gdns.Session      { return nil }
func (w *fakeWriter) Hijack()                     {}

// ParseMsg decodes the DNS message written to this writer.
// WriteTo uses a 2-byte length prefix (TCP framing) before the DNS wire data,
// so we skip those first two bytes before calling Unpack.
func (w *fakeWriter) ParseMsg() (*gdns.Msg, error) {
	b := w.buf.Bytes()
	if len(b) < 2 {
		return nil, fmt.Errorf("response too short: %d bytes", len(b))
	}
	m := new(gdns.Msg)
	m.Data = make([]byte, len(b)-2)
	copy(m.Data, b[2:])
	if err := m.Unpack(); err != nil {
		return nil, err
	}
	return m, nil
}
