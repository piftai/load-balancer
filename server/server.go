package server

import (
	"github.com/piftai/load-balancer/balancer"
	"github.com/piftai/load-balancer/proxy"
	"log"
	"net/http"
	"net/url"
)

// Start is launch load-balancer server
func Start(port string, loadBalancer balancer.Balancer) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		backend := loadBalancer.Next()
		if backend == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("No healthy backends available"))
			return
		}
		target, _ := url.Parse(backend.URL)
		prx := proxy.NewReverseProxy(target)
		prx.ServeHTTP(w, r)
	})

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
