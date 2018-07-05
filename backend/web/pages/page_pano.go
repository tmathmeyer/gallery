package pages

import (
	"../../web"
	"../../database/generated"
	"../../database/util"
	"database/sql"
	"fmt"
)

type Pano struct {}

func (P Pano) PageName() string {
	return "/pano"
}

func (P Pano) TemplateFile() string {
	return "templates/pano_view.html"
}

func (P Pano) TemplateData(db *sql.DB, url []string, Auth web.Authorizer) (interface{}, string, int) {
	if len(url) != 2 {
		return nil, "Not Found", 404
	}

	galleries, err := generated.QueryGalleryTable(db, map[string]interface{}{
		"Path": url[0],
	})

	if err != nil || len(galleries) != 1 {
		return nil, "Not Found", 404
	}

	photos, err := generated.QueryPhotoTable(db, map[string]interface{}{
		"Gallery": galleries[0].Path,
		"Name": url[1],
	})

	if err != nil || len(photos) != 1 {
		return nil, "Not Found", 404
	}

	title := util.GetMetadataValue(db, "siteName")

	return ImageModel {
		Title:            title,
		GalleryName:      galleries[0].Name,
		GalleryPath:      url[0],
		ImagePath:        url[1],
		ImageDescription: photos[0].Description,
		ImageURL:         fmt.Sprintf("/img/%s/%s/O", url[0], url[1]),
	}, "", 200
}