package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ashutos120/go_transfer/internal/config"
	"github.com/ashutos120/go_transfer/internal/utile"
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

func TestUploadWithRemoteImage(t *testing.T) {
	println("uploading remote image")

	resp, err := http.Get("https://static1.cbrimages.com/wordpress/wp-content/uploads/2024/11/a-silent-voice-shoko-1.jpg")
	if err != nil {
		t.Fatal("failed to download image:", err)
	}
	defer resp.Body.Close()

	imgData, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("failed to read image:", err)
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", "image.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = part.Write(imgData); err != nil {
		t.Fatal(err)
	}
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Step 4: Inject fake JWT claim (user ID) into context
	req = req.WithContext(context.WithValue(req.Context(), utile.ClaimsContextKey, "test-user-uuid"))

	// Step 5: Prepare DB and handler
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /upload", Upload)
	handler := withdb(database, withConfig(cfg, mux))

	// Step 6: Execute the request
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Step 7: Check result
	if status := rr.Result().StatusCode; status != http.StatusCreated {
		body, _ := io.ReadAll(rr.Body)
		t.Errorf("upload failed. expected %d, got %d\nBody: %s", http.StatusCreated, status, body)
	} else {
		fmt.Println("Upload successful!")
	}
}
