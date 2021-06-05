package main

import (
	"database/sql"
	
	"github.com/aman7625/iitk-coin/userInfo"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./user_info.db")
	userInfo.CheckError(err)
	user := userInfo.FromSQLite(db)

	user.Add(userInfo.User{
		Rollno: 191000,
		Name:   "Frodo Baggins",
	})
	user.Add(userInfo.User{
		Rollno: 191001,
		Name:   "Bilbo Baggins",
	})
	user.Add(userInfo.User{
		Rollno: 191002,
		Name:   "Samwise Gamgee",
	})

}
