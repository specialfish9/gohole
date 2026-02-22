package dns

import "github.com/specialfish9/confuso/v2"

type Config struct {
	// Upstream is the address of the upstream DNS server to which queries will be forwarded.
	Upstream string `confuso:"upstream" validate:"required"`
	// Address is the address on which the DNS server will listen for incoming queries.
	Address string `confuso:"address" validate:"required"`
	// CacheEnabled toggles the cache. Disabled by default as it is an experimental feature.
	CacheEnabled confuso.Optional[bool] `confuso:"cache"`
}
