package http

import "github.com/specialfish9/confuso/v2"

type Config struct {
	// Address is the address on which the HTTP server will listen for incoming requests.
	Address string `confuso:"address" validate:"required"`
	// ServeFrontend indicates whether to serve the frontend or not.
	ServeFrontend confuso.Optional[bool] `confuso:"serve_frontend"`
}
