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
	"github.com/ashutos120/go_transfer/internal/utile"
)


const configKey = utile.ConfigKey

func withConfig(cfg *config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), configKey, cfg)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {

	//config
	cfg := config.MustConfig()
	//db_connection
	//server
	router := http.NewServeMux()
	//router
	router.HandleFunc("POST /login", Login)
	router.HandleFunc("POST /signup/{username}/{password}", Signup)
	router.HandleFunc("POST /upload", Upload)

	handler := withConfig(cfg, router)

	serverAddr := fmt.Sprintf("%s:%s", cfg.HttpClient.Host, cfg.HttpClient.Port)
	server := &http.Server{
		Addr:    serverAddr,
		Handler: handler,
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
