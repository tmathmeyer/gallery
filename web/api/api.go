package api

import (
	"../../web"
	"database/sql"
	"net/http"
	"strings"
	"../../database/util"
)

type API struct {
	Version  string
	Database *sql.DB
}

type NetReq struct {
	W    http.ResponseWriter
	R    *http.Request
	DB   *sql.DB
	Url  []string
	User string

}

type ObjectHandler interface {
	Get(N NetReq) int
	Post(N NetReq) int
	Put(N NetReq) int
	Delete(N NetReq) int
	Patch(N NetReq) int
	Head(N NetReq) int

	ResourceName() string
}

func GetPathParts(r *http.Request, str string, splitct int) []string {
	restPath := r.URL.Path[len(str):]
	return strings.SplitN(restPath, "/", splitct)
}

func (A API) MakeHandlerFunction(h ObjectHandler, basepath string) http.Handler {
	Authorizer := web.Authorizer{A.Database}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := GetPathParts(r, basepath, -1)
		if len(url) == 1 && url[0] == "" {
			url = []string{}
		}
		req := NetReq{w, r, A.Database, url, Authorizer.GetAuthorization(w, r)}

		switch r.Method {
		case "POST": h.Post(req)
		case "GET": h.Get(req)
		case "PUT": h.Put(req)
		case "DELETE": h.Delete(req)
		case "PATCH": h.Patch(req)
		case "HEAD": h.Head(req)
		}
	})
}

func (A API) AcceptEndpointHandler(h ObjectHandler) {
	fpath := "/api/v" + A.Version + "/" + h.ResourceName() + "/"
	http.Handle(fpath, A.MakeHandlerFunction(h, fpath))
}

func (N NetReq) NotFound() int {
	http.NotFound(N.W, N.R)
	return 404
}

func (N NetReq) Error(msg string, status int) int {
	http.Error(N.W, msg, status)
	return status
}

func (N NetReq) OK() int {
	return N.Error("OK", 200)
}

func (N NetReq) IsAdmin() bool {
	return util.IsUserAdmin(N.DB, N.User)
}

func (N NetReq) Write(msg []byte) {
	N.W.Write(msg)
}

func (N NetReq) ServeFile(f string) int {
	http.ServeFile(N.W, N.R, f)
	return 200
}