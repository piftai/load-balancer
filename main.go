package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/piftai/load-balancer/balancer"
	"github.com/piftai/load-balancer/config"
	"github.com/piftai/load-balancer/repository"
	"github.com/piftai/load-balancer/server"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	db, err := sql.Open("postgres", cfg.Database.Postgres.DSN())
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()
	if err = db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("Successfully connected to database!")

	repo := repository.NewPostgresRepository(db)
	srvr := server.NewServer(repo, cfg)
	srvr.Start(cfg, loadBalancer)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Остановка всех компонентов
	if err = srvr.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	srvr.Limiter.Stop()
	loadBalancer.Stop()

	log.Println("Server exited properly")
}
