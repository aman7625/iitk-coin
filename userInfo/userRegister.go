package userInfo

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Rollno   int    `json:"rollno"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type RegisterResponse struct {
	Message string `json:"message"`
}

func Singup(w http.ResponseWriter,r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status	
		w.WriteHeader(http.StatusBadRequest)
		return 
	}
	
	db, err := sql.Open("sqlite3", "./user_info.db")
	CheckError(err)
	sqldb := FromSQLite(db)
	
	err = sqldb.Add(user)
	if(err != nil) {
		res := RegisterResponse{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(res)
		return
	}
	json.NewEncoder(w).Encode(user)	
}

//GenerateFromPassword does both hashing and salting of password
func HashAndSalt(password string) string {
	pass := []byte(password)
	hashedPassword, err := bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)
    if err != nil {
        panic(err)
    }
	return string(hashedPassword)
}