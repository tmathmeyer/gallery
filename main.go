package main

import (
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"log"
	"net/http"
	"fmt"
	"./web"
	"./web/api"
	"./web/pages"
	"./database/generated"
)

const serverPort string = ":7923"

func setupHandlers(db *sql.DB) {
	// Templated pages
	PAGES := pages.PAGES{db}
	PAGES.AcceptPageHandler(pages.Index{}, false)
	PAGES.AcceptPageHandler(pages.Management{}, true)
	PAGES.AcceptPageHandler(pages.Gallery{}, false)
	PAGES.AcceptPageHandler(pages.Image{}, false)
	PAGES.AcceptPageHandler(pages.Pano{}, false)

	// API endpoints
	// TODO source version from a config file so it is shared in the frontend
	API := api.API{"_dev", db}
	API.AcceptEndpointHandler(api.Gallery{})
	API.AcceptEndpointHandler(api.Users{})
	API.AcceptEndpointHandler(api.Image{})

	// Static files, JS / CSS
	// TODO move to CDN
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Handle login & save cookies to user's browser
	http.Handle("/auth/handle", web.LoginRequestHandler(db))



	// Old Style
	http.Handle("/img/", web.ImageRawHandler(db))
	http.Handle("/conf/", web.CssConfigureHandler(db))
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
