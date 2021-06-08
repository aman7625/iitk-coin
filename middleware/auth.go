package middleware

import (
	"fmt"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
)


func UserAuthentication(w http.ResponseWriter, r *http.Request) {
	// We can obtain the session token from the requests cookies, which come with every request

	c, err := r.Cookie("token")

	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return
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
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Write([]byte(fmt.Sprintf("Welcome %d!", claims.Rollno)))

}