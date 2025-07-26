package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	inmemorystorage "github.com/ashutos120/go_transfer/internal/InMenoryStorge"
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
	localDir := filepath.Join("./uploads", claims)
	done := make(chan error, 1)
	go func() {

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

	fileData := inmemorystorage.Savefile{
		FileName: handler.Filename,
		Path:     filepath.Join(localDir, handler.Filename),
	}
	f, err := fileData.New()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Print("file uuid : ", f.FileID)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("File uploaded successfully\n"))
}

func Download(w http.ResponseWriter, r *http.Request) {
	userID := utile.GetJWTClaim(r)
	if userID == "" {
		http.Error(w, "JWT expired. Please re-login.", http.StatusUnauthorized)
		return
	}
	fileID := r.PathValue("FileID")
	log.Print("[Download] FileID :", fileID)

	if fileID == "" {
		http.Error(w, "Missing FileID", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(fileID)
	if err != nil {
		http.Error(w, "Invalid FileID format", http.StatusBadRequest)
		return
	}

	fileMeta, err := inmemorystorage.GetPath(id)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	file, err := os.Open(fileMeta.Path)
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Detect mime type
	mimeType := mime.TypeByExtension(filepath.Ext(fileMeta.FileName))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Set response headers
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileMeta.FileName))

	// Optional: file size (can help with download progress or browsers)
	if stat, err := file.Stat(); err == nil {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
	}

	// Stream the file
	log.Printf("Download started: user=%s file=%s", userID, fileMeta.FileID.String())

	if _, err := io.Copy(w, file); err != nil {
		log.Printf("Download failed: user=%s file=%s error=%v", userID, fileMeta.FileID.String(), err)
	}
}
