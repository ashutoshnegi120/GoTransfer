package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ashutos120/go_transfer/internal/config"
)

func main() {

	//config
	cfg := config.MustConfig()
	//db_connection
	//server
	router := http.NewServeMux()
	//router
	router.HandleFunc("GET /{$}", Home)
	serverAddr := fmt.Sprintf("%s:%s", cfg.HttpClient.Host, cfg.HttpClient.Port)
	server := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	log.Printf("Server is running at: http://%s", serverAddr)
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-done
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

