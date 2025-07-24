package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

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
	// Step 1: Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max memory
	if err != nil {
		http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Step 2: Retrieve the file from the form
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Step 3: Safely extract JWT claims from context
	claims := utile.GetJWTClaim(r)
	if claims == "" {
		http.Error(w, "Unauthorized: no JWT claims found", http.StatusUnauthorized)
		return
	}

	// Step 4: Log and use the user ID
	log.Printf("user uuid : %s\nUploaded File: %s\n", claims, handler.Filename)

	localDir := "./uploads/" + claims
	err = os.MkdirAll(localDir, os.ModePerm)
	if err != nil {
		http.Error(w, "Failed to create upload directory", http.StatusInternalServerError)
		return
	}

	// Create destination file
	dstPath := fmt.Sprintf("%s/%s", localDir, handler.Filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy uploaded content to destination
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("File uploaded successfully\n"))
}
