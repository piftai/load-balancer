package main

import (
	"github.com/piftai/load-balancer/balancer"
	"github.com/piftai/load-balancer/config"
	"github.com/piftai/load-balancer/server"
)

func main() {
	cfg := config.Load()
	var backends []*balancer.Backend
	for _, b := range cfg.Backends {
		backends = append(backends, &balancer.Backend{
			URL:    b.URL,
			Weight: b.Weight,
			Alive:  true, // Изначально считаем все бэкенды живыми
		})
	}
	loadBalancer := balancer.New(backends, cfg.HealthTick)
	server.Start(cfg, loadBalancer)
}
