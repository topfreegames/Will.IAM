package http

import (
	"net"
	"net/http"
	"time"
)

func getHTTPTransport(cnf *config) http.RoundTripper {
	if _, ok := http.DefaultTransport.(*http.Transport); !ok {
		return http.DefaultTransport // tests use a mock transport
	}
	// We can't get http.DefaultTransport here and update its
	// fields since it's an exported variable, so other libs could
	// also change it and overwrite. This hardcoded values are copied
	// from http.DefaultTransport but could be configurable too.
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          cnf.HTTP.MaxIdleConns,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   cnf.HTTP.MaxIdleConnsPerHost,
	}
}
