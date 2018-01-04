package web

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"../database/generated"
	"../database/util"
)

type CustomImageViewModel struct {
	Title            string
	GalleryName      string
	GalleryPath      string
	ImagePath        string
	ImageDescription string
	ImageURL         string
}

func genericImageViewer(db *sql.DB, path string, templatefile string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		viewdata := GetPathParts(r, path, 2)
		galleryPath := viewdata[0]
		photoName := viewdata[1]

		galleries, err := generated.QueryGalleryTable(db, map[string]interface{}{
			"Path": galleryPath,
		})
		if err != nil || len(galleries) != 1 {
			http.NotFound(w, r)
			return
		}
		gallery := galleries[0]

		photos, err := generated.QueryPhotoTable(db, map[string]interface{}{
			"Gallery": gallery.Path,
			"Name": photoName,
		})
		if err != nil || len(photos) != 1 {
			http.NotFound(w, r)
			return
		}
		photo := photos[0]

		title := util.GetMetadataValue(db, "siteName")

		var Detail = CustomImageViewModel{
			Title:            title,
			GalleryName:      gallery.Name,
			GalleryPath:      galleryPath,
			ImagePath:        photoName,
			ImageDescription: photo.Description,
			ImageURL:         fmt.Sprintf("/img/%s/%s/O", galleryPath, photoName),
		}

		t, _ := template.ParseFiles("templates/"+templatefile)
		t.Execute(w, Detail)
	}) 
}

// GET /view/:gallery/:image
func DragAndDropImageHandler(db *sql.DB) http.Handler {
	return genericImageViewer(db, "/view/", "image.html")
}

// GET /pano/:gallery/:image
func PanoramicImageHandler(db *sql.DB) http.Handler {
	return genericImageViewer(db, "/pano/", "pano_view.html")
}