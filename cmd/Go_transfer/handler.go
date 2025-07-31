package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
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

func Login(w http.ResponseWriter, r *http.Request) {
	database := utile.GetDB(r)
	var uuid_user string
	database.QueryRow(
		"SELECT uuid FROM user WHERE email = ? AND password = ?",
		r.PathValue("email"), r.PathValue("password"),
	).Scan(&uuid_user)
	log.Print("User UUID:", uuid_user)
	parsedUUID, err := uuid.Parse(uuid_user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	secretKey := utile.GetConfig(r).SecretKey
	jwt_token, err := jwt.GenerateJWT(parsedUUID.String(), secretKey)
	if err != nil {
		log.Fatal("jwt generation failed")
		return
	}
	w.WriteHeader(http.StatusAccepted)
	w.Header().Add("Server", "go")
	w.Write([]byte(jwt_token))
}

func Signup(w http.ResponseWriter, r *http.Request) {
	database := utile.GetDB(r)
	// Create table if not exists
	statement, err := database.Prepare(`CREATE TABLE IF NOT EXISTS user (
		uuid TEXT PRIMARY KEY,
		name TEXT,
		password TEXT,
		email TEXT UNIQUE
	)`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err = statement.Exec(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Collect user input
	newUUID := uuid.New()
	u := struct {
		password string
		username string
		email    string
	}{
		password: r.PathValue("password"),
		username: r.PathValue("username"),
		email:    r.PathValue("email"),
	}

	// Check if email already exists
	row := database.QueryRow("SELECT COUNT(*) FROM user WHERE email = ?", u.email)
	var count int
	if err := row.Scan(&count); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if count != 0 {
		http.Error(w, "email already present, use another email", http.StatusBadRequest)
		return
	}

	// Insert into database
	insertStmt, err := database.Prepare("INSERT INTO user (uuid, name, password, email) VALUES (?, ?, ?, ?)")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := insertStmt.Exec(newUUID.String(), u.username, u.password, u.email); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate JWT
	log.Printf("\nUser signed up:\nPassword: %s\nUsername: %s\nEmail: %s", u.password, u.username, u.email)
	secretKey := utile.GetConfig(r).SecretKey
	jwtToken, err := jwt.GenerateJWT(newUUID.String(), secretKey)
	if err != nil {
		log.Fatal("JWT generation failed")
	}

	// Respond with token
	w.WriteHeader(http.StatusCreated)
	w.Header().Add("Server", "Go")
	w.Write([]byte(jwtToken))
}

func Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 100<<20)
	database := utile.GetDB(r)
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
	fileID := uuid.New()
	inmemorystorage.MU.Lock()
	inmemorystorage.FileStatusStore[fileID] = inmemorystorage.StatusProcessing
	inmemorystorage.MU.Unlock()
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
		inmemorystorage.MU.Lock()
		inmemorystorage.FileStatusStore[fileID] = inmemorystorage.StatusFailed
		inmemorystorage.MU.Unlock()
		log.Printf("Failed to save file for user %s: %v", claims, err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	fileData := inmemorystorage.Savefile{
		FileID:   fileID,
		FileName: handler.Filename,
		Path:     filepath.Join(localDir, handler.Filename),
	}
	inmemorystorage.MU.Lock()
	inmemorystorage.FileStatusStore[fileID] = inmemorystorage.StatusUploaded
	inmemorystorage.MU.Unlock()
	stmt, err := database.Prepare(`CREATE TABLE IF NOT EXISTS file (
    UUID TEXT PRIMARY KEY,
    fileName TEXT NOT NULL,
    path TEXT NOT NULL,
    userUUID TEXT NOT NULL
	)`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = stmt.Exec()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert data into table
	stmt, err = database.Prepare("INSERT INTO file (UUID, fileName, path, userUUID) VALUES (?, ?, ?, ?)")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = stmt.Exec(fileData.FileID, fileData.FileName, localDir, claims)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	f, err := fileData.New()
	if err != nil {
		inmemorystorage.MU.Lock()
		inmemorystorage.FileStatusStore[fileID] = inmemorystorage.StatusFailed
		inmemorystorage.MU.Unlock()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Print("file uuid : ", f.FileID)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("File uploaded successfully\n"))
}

func Download(w http.ResponseWriter, r *http.Request) {
	userID := utile.GetJWTClaim(r)
	database := utile.GetDB(r)
	if userID == "" {
		http.Error(w, "JWT expired. Please re-login.", http.StatusUnauthorized)
		return
	}
	fileID := r.PathValue("FileID")
	uuid_parse, err := uuid.Parse(fileID)
	if err != nil {
		http.Error(w, "Missing FileID", http.StatusBadRequest)
		return
	}
	inmemorystorage.MU.Lock()
	inmemorystorage.FileStatusStore[uuid_parse] = inmemorystorage.StatusProcessing
	inmemorystorage.MU.Unlock()
	log.Print("[Download] FileID :", fileID)

	if fileID == "" {
		http.Error(w, "Missing FileID", http.StatusBadRequest)
		return
	}
	var path string
	var Filename string

	fileMeta, err := inmemorystorage.GetPath(uuid_parse)
	if err != nil {
		// Not found in memory, check database
		row := database.QueryRow("SELECT path, fileName FROM file WHERE UUID = ?", uuid_parse)

		var dbPath, dbName string
		if scanErr := row.Scan(&dbPath, &dbName); scanErr != nil {
			if scanErr == sql.ErrNoRows {
				http.Error(w, "Given UUID is not present in the database", http.StatusBadRequest)
				return
			}
			http.Error(w, scanErr.Error(), http.StatusInternalServerError)
			return
		}

		path = filepath.Join(dbPath, dbName)
		Filename = dbName
		dbTOinMO := inmemorystorage.Savefile{
			FileID:   uuid_parse,
			Path:     path,
			FileName: Filename,
		}
		_, err = dbTOinMO.New()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	} else {
		// Found in memory
		path = fileMeta.Path
		Filename = fileMeta.FileName
	}

	file, err := os.Open(path)
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Detect mime type
	mimeType := mime.TypeByExtension(filepath.Ext(Filename))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Set response headers
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, Filename))

	// Optional: file size (can help with download progress or browsers)
	if stat, err := file.Stat(); err == nil {
		inmemorystorage.MU.Lock()
		inmemorystorage.FileStatusStore[uuid_parse] = inmemorystorage.StatusFailed
		inmemorystorage.MU.Unlock()
		w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
	}

	// Stream the file
	log.Printf("Download started: user=%s file=%s", userID, fileID)

	if _, err := io.Copy(w, file); err != nil {
		inmemorystorage.MU.Lock()
		inmemorystorage.FileStatusStore[uuid_parse] = inmemorystorage.StatusFailed
		inmemorystorage.MU.Unlock()
		log.Printf("Download failed: user=%s file=%s error=%v", userID, fileID, err)
	}
	inmemorystorage.MU.Lock()
	inmemorystorage.FileStatusStore[uuid_parse] = inmemorystorage.StatusAvailable
	inmemorystorage.MU.Unlock()
}

func status(w http.ResponseWriter, r *http.Request) {

	claim := utile.GetJWTClaim(r)
	if claim == "" {
		http.Error(w, "not valid id", http.StatusBadRequest)
		return
	}
	fileIDStr := r.PathValue("FileID")
	if fileIDStr == "" {
		http.Error(w, "Please provide a valid fileID", http.StatusBadRequest)
		return
	}

	fileID, err := uuid.Parse(fileIDStr)
	if err != nil {
		http.Error(w, "Invalid fileID format", http.StatusBadRequest)
		return
	}

	inmemorystorage.MU.Lock()
	status, ok := inmemorystorage.FileStatusStore[fileID]
	inmemorystorage.MU.Unlock()

	if !ok {
		http.Error(w, "FileID not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"file_id": fileID.String(),
		"status":  string(status),
	})
}
