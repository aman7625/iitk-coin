package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
)

type Response struct {
	Message string `json:"message"`
}

func UserAuthentication(w http.ResponseWriter, r *http.Request) (int64, error) {
	// We can obtain the session token from the requests cookies, which come with every request
	var ret int64
	c, err := r.Cookie("token")

	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			return ret, err
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return ret, err
	}

	// Get the JWT string from the cookie
	tknStr := c.Value

	jwtWrapper := JwtWrapper{
		SecretKey:      "my_secret_key",
		Issuer:         "AuthService",
		ExpirationMins: 10,
	}

	claims, err := jwtWrapper.ValidateToken(tknStr)
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return ret, err
		}
		w.WriteHeader(http.StatusBadRequest)
		return ret, err
	}

	return claims.Rollno, nil
	//w.Write([]byte(fmt.Sprintf("Welcome %d!", claims.Rollno)))

}

func Welcome(w http.ResponseWriter, r *http.Request) {
	rollno, err := UserAuthentication(w, r)

	if err != nil {
		res := Response{
			Message: err.Error(),
		}
		json.NewEncoder(w).Encode(res)
		return
	}

	w.Write([]byte(fmt.Sprintf("Welcome %d!", rollno)))
}
