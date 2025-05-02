package server

import (
	"context"
	"github.com/piftai/load-balancer/balancer"
	"github.com/piftai/load-balancer/config"
	"github.com/piftai/load-balancer/proxy"
	"github.com/piftai/load-balancer/ratelimiter"
	"github.com/piftai/load-balancer/repository"
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

type Server struct {
	clientHandler ClientHandler
	Limiter       *ratelimiter.ClientLimiter
	server        *http.Server
}

func NewServer(repo repository.Repository, cfg config.Config) *Server {
	return &Server{
		clientHandler: *NewClientHandler(repo),
		Limiter:       ratelimiter.NewClientLimiter(cfg.RateLimiting.DefaultCapacity, cfg.RateLimiting.DefaultRate),
	}
}

// Start is launch load-balancer server
func (s *Server) Start(cfg config.Config, loadBalancer Balancer) {
	limiter := ratelimiter.NewClientLimiter(cfg.RateLimiting.DefaultCapacity, cfg.RateLimiting.DefaultRate)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /clients", s.clientHandler.Create)
	mux.HandleFunc("GET /clients/{id}", s.clientHandler.Get)
	mux.HandleFunc("PUT /clients", s.clientHandler.Update)
	mux.HandleFunc("DELETE /clients/{id}", s.clientHandler.Delete)
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

func (s *Server) Shutdown(ctx context.Context) error {
	return s.Shutdown(ctx)
}
