package main

import (
	"database/sql"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"time"
)

func UnauthorizedFailPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unauthorized", 403)
	})
}

func verify_authentication_middleware(authorized http.Handler, unauthorized http.Handler, db *sql.DB) http.Handler {
	if unauthorized == nil {
		unauthorized = UnauthorizedFailPage()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authentication := r.Header.Get("Authorization")
		if authentication == "" {
			authcookie, err := r.Cookie("jwt")
			if err != nil {
				fmt.Printf("JWT: %s\n", err)
				unauthorized.ServeHTTP(w, r)
				return
			}
			authentication = authcookie.Value
		}

		if authentication == "" {
			fmt.Printf("AUTH: missing\n")
			unauthorized.ServeHTTP(w, r)
			return
		} else if authentication[:6] == "Bearer" {
			authentication = authentication[7:]
		}

		secret := getMetadataValue(db, "secret")
		_, err := get_user_authentication(authentication, []byte(secret))

		if err != nil {
			fmt.Printf("PCMP: %s\n", err)
			unauthorized.ServeHTTP(w, r)
		} else {
			authorized.ServeHTTP(w, r)
		}
	})
}

func get_user_authentication(bearer_token string, secret []byte) (string, error) {
	token, err := jwt.Parse(bearer_token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return "", fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["user"].(string), nil
	} else {
		return "", err
	}
}

func get_authentication_token(secret []byte, user string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": user,
		"nbf":  time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
	})

	return token.SignedString(secret)
}
