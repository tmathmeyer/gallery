package api

import (
	"net/http"
	"database/sql"
	"strings"
	"../../web"
)

type API struct {
	Version string
}

type ObjectHandler interface {
	Get(w http.ResponseWriter, r *http.Request, url []string)
	Post(w http.ResponseWriter, r *http.Request, url []string)
	Put(w http.ResponseWriter, r *http.Request, url []string)
	Delete(w http.ResponseWriter, r *http.Request, url []string)
	Patch(w http.ResponseWriter, r *http.Request, url []string)
	Head(w http.ResponseWriter, r *http.Request, url []string)

	ResourceName() string
	GetDatabase() *sql.DB
}


func GetPathParts(r *http.Request, str string, splitct int) []string {
	restPath := r.URL.Path[len(str):]
	return strings.SplitN(restPath, "/", splitct)
}

func (A API) MakeHandlerFunction(h ObjectHandler, path string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := GetPathParts(r, path, -1);
		if len(url) == 1 && url[0] == "" {
			url = []string{}
		}
		switch r.Method {
		case "POST":
			h.Post(w, r, url);
		case "GET":
			h.Get(w, r, url);
		case "PUT":
			h.Put(w, r, url);
		case "DELETE":
			h.Delete(w, r, url);
		case "PATCH":
			h.Patch(w, r, url);
		case "HEAD":
			h.Head(w, r, url);
		}
	})
}

func (A API) AcceptEndpointHandler(h ObjectHandler) {
	fpath := "/api/v" + A.Version + "/" + h.ResourceName() + "/"
	http.Handle(fpath, A.MakeHandlerFunction(h, fpath))
}

func (A API) AcceptEndpointHandlerAuthenticated(h ObjectHandler) {
	fpath := "/api/v" + A.Version + "/" + h.ResourceName() + "/"
	http.Handle(fpath, web.VerifyAuthenticationMiddleware(
		A.MakeHandlerFunction(h, fpath), nil, h.GetDatabase()))
}
