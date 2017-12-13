package web

import (
	"database/sql"
	"html/template"
	"net/http"
	"../database/generated"
	"../database/util"
	"log"
)

type IndexModel struct {
	Title     string
	APIKey    string
	Galleries []generated.Gallery
}

// GET /
func IndexHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t, _ := template.ParseFiles("templates/index.html")

		galleries, err := generated.QueryGalleryTable(db, map[string]interface{}{})
		if err != nil {
			log.Fatal(err)
			http.NotFound(w, r)
			return
		}

		title := util.GetMetadataValue(db, "siteName")
		apiKey := util.GetMetadataValue(db, "gmapsApiKey")

		var index = IndexModel{
			Title:     title,
			APIKey:    apiKey,
			Galleries: galleries}
		t.Execute(w, index)
	})
}