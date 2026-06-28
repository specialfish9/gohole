package dns

import "github.com/specialfish9/confuso/v2"

type BlockingStrategy = string

const (
	// BlockingStrategyNXDOMAIN returns NXDOMAIN for blocked queries.
	BlockingStrategyNXDOMAIN BlockingStrategy = "nxdomain"
	// BlockingStrategyIP returns a dumb IP address for blocked queries (usually, 0.0.0.0 or ::).
	BlockingStrategyIP BlockingStrategy = "ip"
)

type Config struct {
	// Upstream is the address of the upstream DNS server to which queries will be forwarded.
	Upstream string `confuso:"upstream"          validate:"required"`
	// Address is the address on which the DNS server will listen for incoming queries.
	Address string `confuso:"address"           validate:"required"`
	// CacheEnabled toggles the cache. Disabled by default as it is an experimental feature.
	CacheEnabled  confuso.Optional[bool]           `confuso:"cache"`
	CustomDomains confuso.Optional[map[string]any] `confuso:"custom_domains"`
	// BlockingStrategy defines how blocked queries are handled. Default is "nxdomain".
	BlockingStrategy confuso.Optional[BlockingStrategy] `confuso:"blocking_strategy"`
}
