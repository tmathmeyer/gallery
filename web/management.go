package web

import (
	"database/sql"
	"html/template"
	"net/http"


	"../database/util"
	"../database/generated"
)

func GalleryManagementHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t, _ := template.ParseFiles("templates/manage.html")

		galleries, err := generated.QueryGalleryTable(db, map[string]interface{}{})
		if err != nil {
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