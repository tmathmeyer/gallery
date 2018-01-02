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

/* A pretty standard GTFO message */
func UnauthorizedFailPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unauthorized", 403)
	})
}

func LoginPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/login.html")
	})
}

func RedirectToManagement() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/manage", 302)
	})
}

/* A middleware for checking authentication and calling a success or deny handler */
func VerifyAuthenticationMiddleware(authorized http.Handler, unauthorized http.Handler, db *sql.DB) http.Handler {
	if unauthorized == nil {
		unauthorized = UnauthorizedFailPage()
	}

	if authorized == nil {
		log.Fatal("Success handler required!")
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authentication := r.Header.Get("Authorization")
		if authentication == "" {
			authcookie, err := r.Cookie("jwt")
			if err != nil {
				unauthorized.ServeHTTP(w, r)
				return
			}
			authentication = authcookie.Value
		}

		if authentication == "" {
			unauthorized.ServeHTTP(w, r)
			return
		} else if authentication[:6] == "Bearer" {
			authentication = authentication[7:]
		}

		secret := util.GetMetadataValue(db, "secret")
		user, err := GetUserAuthentication(authentication, []byte(secret))

		if err != nil {
			unauthorized.ServeHTTP(w, r)
		} else {
			is_admin := "user"
			if util.IsUserAdmin(db, user) {
				is_admin = "admin"
			}
			admin := http.Cookie{Name:"admin", Value:is_admin}
			cookie := http.Cookie{Name:"username", Value:user}
			r.AddCookie(&cookie)
			r.AddCookie(&admin)
			authorized.ServeHTTP(w, r)
		}
	})
}

func GetUserAuthentication(bearer_token string, secret []byte) (string, error) {
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