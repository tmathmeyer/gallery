package web

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"../database/util"
	"../database/generated"
)

type GalleryDetailModel struct {
	GalleryName string
	GalleryID   string
	Images      []generated.Photo
	Title       string
	GpxPresent  bool
	APIKey      string
}

// GET /gallery/:gallery
func GalleryDetailhandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		galleryName := GetPathParts(r, "/gallery/", 1)[0]

		galleries, err := generated.QueryGalleryTable(db, map[string]interface{}{
			"Path": galleryName,
		})

		if err != nil || len(galleries) != 1 {
			http.NotFound(w, r)
			return
		}

		gallery := galleries[0]

		photos, err := generated.QueryPhotoTable(db, map[string]interface{}{
			"Gallery": gallery.Path,
		})
		if err != nil {
			http.NotFound(w, r)
			return
		}

		title := util.GetMetadataValue(db, "siteName")
		apiKey := util.GetMetadataValue(db, "gmapsApiKey")
		dataFsLocation := util.GetMetadataValue(db, "dataStore")

		resource := fmt.Sprintf("%s/%d/route.gpx", dataFsLocation, gallery.Id)
		gpxPresent := FExists(resource)

		var Detail = GalleryDetailModel{
			GalleryName: gallery.Name,
			GalleryID:   gallery.Path,
			Images:      photos,
			Title:       title,
			GpxPresent:  gpxPresent,
			APIKey:      apiKey,
		}

		t, _ := template.ParseFiles("templates/detail.html")
		t.Execute(w, Detail)
		return
	})
}
