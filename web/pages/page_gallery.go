package pages

import (
	"../../web"
	"../../database/generated"
	"../../database/util"
	"database/sql"
	"fmt"
)

type Gallery struct {}

type GalleryModel struct {
	GalleryName string
	GalleryID   string
	Images      []generated.Photo
	Title       string
	GpxPresent  bool
	APIKey      string
}

func (G Gallery) PageName() string {
	return "/gallery"
}

func (G Gallery) TemplateFile() string {
	return "templates/detail.html"
}

func (G Gallery) TemplateData(db *sql.DB, url []string, Auth web.Authorizer) (interface{}, string, int) {
	galleries, err := generated.QueryGalleryTable(db, map[string]interface{}{
		"Path": url[0],
	})

	if err != nil || len(galleries) != 1 {
		return nil, "Not Found", 404
	}

	photos, err := generated.QueryPhotoTable(db, map[string]interface{}{
		"Gallery": galleries[0].Path,
	})

	if err != nil {
		return nil, "Not Found", 404
	}

	title := util.GetMetadataValue(db, "siteName")
	apiKey := util.GetMetadataValue(db, "gmapsApiKey")
	dataFsLocation := util.GetMetadataValue(db, "dataStore")

	resource := fmt.Sprintf("%s/%d/route.gpx", dataFsLocation, galleries[0].Id)
	gpxPresent := web.FExists(resource)

	return GalleryModel {
		GalleryName: galleries[0].Name,
		GalleryID:   galleries[0].Path,
		Images:      photos,
		Title:       title,
		GpxPresent:  gpxPresent,
		APIKey:      apiKey,
	}, "", 200
}