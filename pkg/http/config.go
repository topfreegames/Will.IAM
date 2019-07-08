package http

import (
	"net/http"
	"time"
)

type (
	config struct {
		HTTP       *configHTTP
		URL        string
		Middleware *configMiddleware
		Permission *configPermission
	}

	configHTTP struct {
		MaxIdleConns        int
		MaxIdleConnsPerHost int
		Timeout             time.Duration
	}

	configMiddleware struct {
		Enabled bool
	}

	configPermission struct {
		Service string
	}
)

// NewConfig returns the config struct with default values
func NewConfig() *config {
	return &config{
		HTTP: &configHTTP{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: http.DefaultMaxIdleConnsPerHost,
			Timeout:             500 * time.Millisecond,
		},
		URL: "http://localhost:4040",
		Middleware: &configMiddleware{
			Enabled: false,
		},
		Permission: &configPermission{
			Service: "service",
		},
	}
}
