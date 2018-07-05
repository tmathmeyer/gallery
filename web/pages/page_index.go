package pages

import (
	"../../web"
	"../../database/generated"
	"../../database/util"
	"database/sql"
)

type Index struct {}
type IndexModel struct {
	Title string
	APIKey string
	Galleries []generated.Gallery
}

func (I Index) PageName() string {
	return ""
}

func (I Index) TemplateFile() string {
	return "templates/index.html"
}

func (I Index) TemplateData(db *sql.DB, url []string, Auth web.Authorizer) (interface{}, string, int) {
	galleries, err := generated.QueryGalleryTable(db, map[string]interface{}{})
	if err != nil {
		return nil, "Could not load galleries", 500
	}

	title := util.GetMetadataValue(db, "siteName")
	apiKey := util.GetMetadataValue(db, "gmapsApiKey")

	return IndexModel {
		Title:     title,
		APIKey:    apiKey,
		Galleries: galleries,
	}, "", 200
}