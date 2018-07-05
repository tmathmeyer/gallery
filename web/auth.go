package web

import (
	"database/sql"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"time"
	"../database/util"
	"log"
	"fmt"
)

type Authorizer struct {
	DB *sql.DB
}



func (A Authorizer) GetUserAuthentication(bearer_token string, secret []byte) (string, error) {
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

func (A Authorizer) GetUserFromAuthorization(authtoken string) string {
	secret := util.GetMetadataValue(A.DB, "secret")
	user, err := A.GetUserAuthentication(authtoken, []byte(secret))
	if err != nil {
		return ""
	}
	return user
}

func (A Authorizer) GetAuthorization(w http.ResponseWriter, r *http.Request) string {
	authorization := r.Header.Get("Authorization")
	if authorization != "" {
		if authorization[:6] == "Bearer" {
			authorization = authorization[7:]
		}
		return A.GetUserFromAuthorization(authorization)
	}

	cookie, err := r.Cookie("jwt")
	if err == nil {
		return A.GetUserFromAuthorization(cookie.Value)
	}

	return "";
}

func (A Authorizer) Middleware(authorized http.Handler, unauthorized http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := A.GetAuthorization(w, r)
		if user != "" && util.IsUserAdmin(A.DB, user) {
			authorized.ServeHTTP(w, r)
			return
		}

		unauthorized.ServeHTTP(w, r)
	})
}






func RedirectToManagement() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/manage", 302)
	})
}

func GeneratetAuthenticationToken(secret []byte, user string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": user,
		"nbf":  time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
	})

	return token.SignedString(secret)
}

func LoginRequestHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		user_a := r.Form["username"]
		pass_a := r.Form["password"]
		if len(user_a) != 1 || len(pass_a) != 1 {
			http.Error(w, "Unauthorized", 400)
			return
		}
		user := user_a[0]
		pass := pass_a[0]

		if util.ValidCredentials(db, user, pass) {
			secret := util.GetMetadataValue(db, "secret")
			token, err := GeneratetAuthenticationToken([]byte(secret), user)

			if err != nil {
				http.Error(w, "Unauthorized", 400)
				return
			}

			expiration := time.Now().Add(365 * 24 * time.Hour)
			cookie := http.Cookie{Name: "jwt", Value: token, Expires: expiration, Path: "/"}
			http.SetCookie(w, &cookie)
			RedirectToManagement().ServeHTTP(w, r)
		} else {
			log.Println(fmt.Sprintf("Failed login user:{%s} password:{%s}", user, pass))
			http.Error(w, "Unauthorized", 400)
		}
	})
}