package server

import (
	"github.com/piftai/load-balancer/balancer"
	"github.com/piftai/load-balancer/config"
	"github.com/piftai/load-balancer/proxy"
	"github.com/piftai/load-balancer/ratelimiter"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Balancer interface {
	Next() *balancer.Backend
	HealthChecker(interval time.Duration)
}

// Start is launch load-balancer server
func Start(cfg config.Config, loadBalancer Balancer) {
	limiter := ratelimiter.NewClientLimiter(cfg.RateLimiting.DefaultCapacity, time.Duration(cfg.RateLimiting.DefaultRate*1000))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		clientID := r.Header.Get("\"X-Client-ID\"")
		if clientID == "" {
			clientID = strings.Split(r.RemoteAddr, ":")[0]
		}
		if !limiter.Allow(clientID) {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("No tokens"))
			log.Println("Dont have tokens left clientID:", clientID)
			return
		}
		backend := loadBalancer.Next()
		if backend == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("No healthy backends available"))
			return
		}
		log.Println("backend catched:", backend.URL)
		target, _ := url.Parse(backend.URL)
		prx := proxy.NewReverseProxy(target)
		prx.ServeHTTP(w, r)
	})

	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
