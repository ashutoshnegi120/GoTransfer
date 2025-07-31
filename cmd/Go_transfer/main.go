package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ashutos120/go_transfer/internal/config"
	"github.com/ashutos120/go_transfer/internal/midlleware"
	"github.com/ashutos120/go_transfer/internal/utile"
	_ "github.com/mattn/go-sqlite3"
)

const configKey = utile.ConfigKey

func withConfig(cfg *config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), configKey, cfg)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func withdb(db *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db := context.WithValue(r.Context(), utile.DbKey, db)
		next.ServeHTTP(w, r.WithContext(db))
	})
}

func main() {

	cfg := config.MustConfig()
	database, err := sql.Open("sqlite3", cfg.DatabaseLocation)
	if err != nil {
		log.Fatal("error :", err.Error())
	}
	router := http.NewServeMux()
	router.HandleFunc("POST /login/{password}/{email}", Login)
	router.HandleFunc("POST /signup/{username}/{password}/{email}", Signup)
	router.HandleFunc("POST /upload", midlleware.JWTMiddleware(cfg.SecretKey, Upload))
	router.HandleFunc("GET /download/{FileID}", midlleware.JWTMiddleware(cfg.SecretKey, Download))
	router.HandleFunc("GET /stats/{FileID}", midlleware.JWTMiddleware(cfg.SecretKey, status))

	handler := withConfig(cfg, router)
	handlerwithdb := withdb(database, handler)

	serverAddr := fmt.Sprintf("%s:%s", cfg.HttpClient.Host, cfg.HttpClient.Port)
	server := &http.Server{
		Addr:    serverAddr,
		Handler: handlerwithdb,
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
	defer database.Close()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
