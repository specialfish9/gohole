package dns

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"codeberg.org/miekg/dns"
)

type DohClient struct {
	httpClient *http.Client
}

func NewDohClient() *DohClient {
	return &DohClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *DohClient) Exchange(
	ctx context.Context,
	m *dns.Msg,
	network, address string,
) (*dns.Msg, time.Duration, error) {
	start := time.Now()

	if err := m.Pack(); err != nil {
		return nil, 0, fmt.Errorf("doh client: packing message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, address, bytes.NewReader(m.Data))
	if err != nil {
		return nil, 0, fmt.Errorf("doh client: creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/dns-message")
	req.Header.Set("Accept", "application/dns-message")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("doh client: performing http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("doh client: bad status code %d", resp.StatusCode)
	}

	respBuf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("doh client: reading response body: %w", err)
	}

	respMsg := new(dns.Msg)
	respMsg.Data = respBuf
	if err := respMsg.Unpack(); err != nil {
		return nil, 0, fmt.Errorf("doh client: unpacking response: %w", err)
	}

	if respMsg.ID != m.ID {
		return nil, 0, fmt.Errorf(
			"doh client: response ID %d does not match request ID %d",
			respMsg.ID,
			m.ID,
		)
	}

	return respMsg, time.Since(start), nil
}
