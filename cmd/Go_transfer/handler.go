package main

import (
	"log"
	"net/http"

	"github.com/ashutos120/go_transfer/internal/jwt"
	"github.com/google/uuid"
)

type user struct {
	username string
	password string
}

func Login(w http.ResponseWriter, r *http.Request) {
	uuid := uuid.New()
	secretKey := GetConfig(r).SecretKey
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
	secretKey := GetConfig(r).SecretKey
	jwt_token, err := jwt.GenerateJWT(uuid.String(), secretKey)
	if err != nil {
		log.Fatal("jwt generation failed .........")
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Add("server", "go")
	w.Write([]byte(jwt_token))

}
