package main

import (
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"log"
	"net/http"
	"fmt"
	"./web"
	"./web/api"
	"./database/generated"
)

const serverPort string = ":7923"

func setupHandlers(db *sql.DB) {
	// No authentication
	http.Handle("/", web.IndexHandler(db))
	http.Handle("/gallery/", web.GalleryDetailhandler(db))
	http.Handle("/gdata/", web.GalleryDataHandler(db))
	http.Handle("/view/", web.DragAndDropImageHandler(db))
	http.Handle("/pano/", web.PanoramicImageHandler(db))
	http.Handle("/img/", web.ImageRawHandler(db))

	http.Handle("/conf/", web.CssConfigureHandler(db))

	// Related to Authentication
	http.Handle("/auth/handle", web.LoginRequestHandler(db))
	http.Handle("/manage", web.VerifyAuthenticationMiddleware(
							web.GalleryManagementHandler(db),
							web.LoginPage(), db))
	http.Handle("/api/v0/", web.VerifyAuthenticationMiddleware(web.ApiHandler(db), nil, db))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	API := api.API{"_dev", db}
	API.AcceptEndpointHandler(api.Gallery{})
	API.AcceptEndpointHandler(api.Users{})
}

func main() {
	db, err := generated.OpenDatabase("live.sqlite")
	if err != nil {
		log.Fatal(err)
	}

	setupHandlers(db)

	fmt.Printf("Starting gallery server on port %s\n", serverPort)
	http.ListenAndServe(serverPort, nil)
}
