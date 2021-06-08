package main

import (
	"log"
	"net/http"

	"github.com/aman7625/iitk-coin/middleware"
	"github.com/aman7625/iitk-coin/userInfo"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	
	r := mux.NewRouter();

	r.HandleFunc("/login",userInfo.Login).Methods("POST")
	r.HandleFunc("/signup",userInfo.Singup).Methods("POST")
	r.HandleFunc("/secretpage",middleware.UserAuthentication).Methods("GET")

	log.Println("Server Starting...!")
	log.Fatal(http.ListenAndServe(":8000",r))
	
}

/*
	//To check users in database
	db, err := sql.Open("sqlite3", "./user_info.db")
	userInfo.CheckError(err)
	user := userInfo.FromSQLite(db)
	users := user.Get();
	fmt.Println(users)
	*/