package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ashutos120/go_transfer/internal/config"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

var (
	cfg *config.Config // assuming your config.MustConfig returns *Config
)

func TestMain(m *testing.M) {
	cfg = config.MustConfig() // Pass the path
	code := m.Run()
	os.Exit(code)
}

func TestSignup(t *testing.T) {
	println("testing signup")
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Error("error:", err)
	}
	defer database.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /signup/{username}/{password}/{email}", Signup)
	handler := withdb(database, withConfig(cfg, mux))

	_, err = database.Exec(`
        CREATE TABLE IF NOT EXISTS user (
            uuid TEXT PRIMARY KEY ,
            name TEXT,
            password TEXT ,
            email TEXT UNIQUE
        )
    `)
	if err != nil {
		t.Fatal("error creating user table:", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/signup/atul/123321/atulnegisgrr%40gmail.com", nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Result().StatusCode; status != http.StatusCreated {
		t.Errorf("wrong status code. expected %d, got %d", http.StatusCreated, status)
	}

	if jwt := rr.Body.String(); len(jwt) == 0 {
		t.Error("expected JWT token in response body, got empty response")
	}
}

func TestLogin(t *testing.T) {
	println("testing Login")
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Error("error:", err)
	}
	defer database.Close()

	_, err = database.Exec(`
        CREATE TABLE IF NOT EXISTS user (
            uuid TEXT PRIMARY KEY ,
            name TEXT,
            password TEXT ,
            email TEXT UNIQUE
        )
    `)
	if err != nil {
		t.Fatal("error creating users table:", err)
	}

	_, err = database.Exec(`
    INSERT INTO user (uuid, name, password, email)
    VALUES (?, ?, ?, ?)`, uuid.New().String(), "atul", "123321", "atulnegisgrr@gmail.com")

	if err != nil {
		t.Fatal("error inserting test user:", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /login/{password}/{email}", Login)
	handler := withdb(database, withConfig(cfg, mux))

	req := httptest.NewRequest(http.MethodPost, "/login/123321/atulnegisgrr%40gmail.com", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Result().StatusCode; status != http.StatusAccepted {
		t.Errorf("wrong status code. expected %d, got %d", http.StatusAccepted, status)
	}

	if jwt := rr.Body.String(); len(jwt) == 0 {
		t.Error("expected JWT token in response body, got empty response")
	}
}
