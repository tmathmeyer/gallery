package pages

import (
	"../../web"
	"../../database/generated"
	"../../database/util"
	"database/sql"
)

type Management struct {}
type ManagementModel struct {
	Title string
	APIKey string
	Galleries []generated.Gallery
}

func (M Management) PageName() string {
	return "/manage"
}

func (M Management) TemplateFile() string {
	return "templates/manage.html"
}

func (M Management) TemplateData(db *sql.DB, url []string, Auth web.Authorizer) (interface{}, string, int) {
	galleries, err := generated.QueryGalleryTable(db, map[string]interface{}{})
	if err != nil {
		return nil, "Could not load galleries", 500
	}

	title := util.GetMetadataValue(db, "siteName")
	apiKey := util.GetMetadataValue(db, "gmapsApiKey")

	return ManagementModel {
		Title:     title,
		APIKey:    apiKey,
		Galleries: galleries,
	}, "", 200
}