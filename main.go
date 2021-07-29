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

	r := mux.NewRouter()

	r.HandleFunc("/login", userInfo.Login).Methods("POST")
	r.HandleFunc("/signup", userInfo.Singup).Methods("POST")
	r.HandleFunc("/secretpage", middleware.Welcome).Methods("GET")

	r.HandleFunc("/reward", userInfo.AwardCoins).Methods("POST")
	r.HandleFunc("/transfer", userInfo.TransferCoins).Methods("POST")
	r.HandleFunc("/view", userInfo.CoinBalance).Methods("GET")
	r.HandleFunc("/redeem", userInfo.RedeemCoins).Methods("POST")
	r.HandleFunc("/takeAction", userInfo.TakeAction).Methods("POST")
	r.HandleFunc("/destroy", userInfo.DestroyGraduatingBatchAccounts).Methods("POST")

	log.Println("Starting Server...!")
	log.Fatal(http.ListenAndServe(":8080", r))

}

/*
	//To check users in database
	db, err := sql.Open("sqlite3", "./user_info.db")
	userInfo.CheckError(err)
	user := userInfo.FromSQLite(db)
	users := user.Get();
	fmt.Println(users)
*/
