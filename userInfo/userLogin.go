package userInfo

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/aman7625/iitk-coin/middleware"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Rollno   int64  `json:"rollno"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user LoginRequest

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite3", "./user_info.db")
	CheckError(err)
	s := FromSQLite(db)

	rollno := user.Rollno
	password := user.Password

	var dbpass string
	err = s.DB.QueryRow("SELECT password FROM user_info WHERE rollno=?", rollno).Scan(&dbpass)
	if err != nil {
		// If an entry with the Roll Number does not exist, send an "Unauthorized"(401) status
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// If the error is of any other type, send a 500 status
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbpass), []byte(password))
	if err != nil {
		// If the two passwords don't match, return a 401 status
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	jwtWrapper := middleware.JwtWrapper{
		SecretKey:      "my_secret_key",
		Issuer:         "AuthService",
		ExpirationMins: 10,
	}

	signedToken, err := jwtWrapper.GenerateToken(user.Rollno)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	tokenResponse := LoginResponse{
		Token: signedToken,
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   signedToken,
		Expires: time.Now().Local().Add(time.Minute * time.Duration(jwtWrapper.ExpirationMins)),
		Path:    "/",
	})
	json.NewEncoder(w).Encode(tokenResponse)
}
