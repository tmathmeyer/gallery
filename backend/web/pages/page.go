package pages

import (
	"../../web"
	"database/sql"
	"net/http"
	"html/template"
)

type PAGES struct {
	Database *sql.DB
}

type PageGenerator interface {
	PageName() string
	TemplateFile() string
	TemplateData(db *sql.DB, url []string, Auth web.Authorizer) (interface{}, string, int)
}

func (P PAGES) AcceptPageHandler(p PageGenerator, admin_only bool) {
	authorizer := web.Authorizer{P.Database}
	handler := P.MakeHandlerFunction(p, authorizer)

	if admin_only {
		handler = authorizer.Middleware(handler, P.MakeHandlerFunction(Login{}, authorizer))
	}

	http.Handle(p.PageName() + "/", handler)
}

func (P PAGES) MakeHandlerFunction(p PageGenerator, a web.Authorizer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Not Found", 404)
			return
		}

		url := web.GetPathParts(r, "/"+p.PageName(), -1)
		if len(url) == 1 && url[0] == "" {
			url = []string{}
		}

		t, err := template.ParseFiles(p.TemplateFile())
		if err != nil {
			http.Error(w, "Not Found", 404)
			return
		}

		data, msg, code := p.TemplateData(P.Database, url, a)
		if code != 200 {
			http.Error(w, msg, code)
		}

		t.Execute(w, data)
	})
}