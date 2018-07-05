package pages

import (
	"../../web"
	"database/sql"
)

type Login struct {}

func (L Login) PageName() string {
	return "/login"
}

func (L Login) TemplateFile() string {
	return "templates/login.html"
}

func (L Login) TemplateData(db *sql.DB, url []string, Auth web.Authorizer) (interface{}, string, int) {
	return 1, "", 200
}