package proxy

import (
	"net/http/httputil"
	"net/url"
)

func NewReverseProxy(target *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	return proxy
}
