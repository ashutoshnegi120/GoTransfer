package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ashutos120/go_transfer/internal/jwt"
	"github.com/ashutos120/go_transfer/internal/utile"
	"github.com/google/uuid"
)

type user struct {
	username string
	password string
}

func Login(w http.ResponseWriter, r *http.Request) {
	uuid := uuid.New()
	secretKey := utile.GetConfig(r).SecretKey
	jwt_token, err := jwt.GenerateJWT(uuid.String(), secretKey)
	if err != nil {
		log.Fatal("jwt generation failed .........")
	}
	w.WriteHeader(http.StatusAccepted)
	w.Header().Add("server", "go")
	w.Write([]byte(jwt_token))
}

func Signup(w http.ResponseWriter, r *http.Request) {
	uuid := uuid.New()
	user := user{
		password: r.PathValue("password"),
		username: r.PathValue("username"),
	}
	log.Printf("\nuser : \n password : %s \n username : %s ", user.password, user.username)
	secretKey := utile.GetConfig(r).SecretKey
	jwt_token, err := jwt.GenerateJWT(uuid.String(), secretKey)
	if err != nil {
		log.Fatal("jwt generation failed .........")
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Add("server", "go")
	w.Write([]byte(jwt_token))

}

func Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 100<<20)

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	claims := utile.GetJWTClaim(r)
	if claims == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// ✅ Step 1: Copy file into memory (or temp file if large)
	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)
	if err != nil {
		http.Error(w, "Failed to buffer file", http.StatusInternalServerError)
		return
	}

	// ✅ Step 2: Save asynchronously
	done := make(chan error, 1)
	go func() {
		localDir := filepath.Join("./uploads", claims)
		if err := os.MkdirAll(localDir, os.ModePerm); err != nil {
			done <- err
			return
		}

		dstPath := filepath.Join(localDir, handler.Filename)
		dst, err := os.Create(dstPath)
		if err != nil {
			done <- err
			return
		}
		defer dst.Close()

		_, err = io.Copy(dst, &buf)
		done <- err
	}()

	// ✅ Step 3: Wait and respond
	if err := <-done; err != nil {
		log.Printf("Failed to save file for user %s: %v", claims, err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("File uploaded successfully\n"))
}
