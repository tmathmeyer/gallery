package api

import (
	"encoding/json"
	"../../database/generated"
	"../../database/util"
)

type Users struct {}

func (G Users) Get(N NetReq) int {
	if !N.IsAdmin() {
		return N.Error("Unauthorized", 403)
	}
	users, err := generated.QueryUserTable(N.DB, map[string]interface{}{})
	for i := range users {
		users[i].Passhash = "[redacted]"
	}
	jData, err := json.Marshal(users)
	if err != nil {
		return N.Error("Could not provide list", 500)
	}

	N.W.Header().Set("Content-Type", "application/json")
	N.Write(jData)
	return 200
}

func (G Users) Post(N NetReq) int {
	if !N.IsAdmin() {
		return N.Error("Unauthorized", 403)
	}
	N.R.ParseMultipartForm(32 << 20)
	util.AddUser(N.DB, N.R.Form["username"][0], N.R.Form["password"][0])
	return N.OK()
}

func (G Users) Put(N NetReq) int {
	if !N.IsAdmin() {
		return N.Error("Unauthorized", 403)
	}
	N.R.ParseMultipartForm(32 << 20)
	uidl := N.R.Form["id"]
	if len(uidl) == 1 && util.ChangePasswordForId(N.DB, uidl[0], N.R.Form["password"][0]) {
		return N.OK()
	}
	if len(uidl) != 1 && util.ChangePassword(N.DB, N.User, N.R.Form["password"][0]) {
		return N.OK()
	} 
	return N.Error("Failed", 500)
}

func (G Users) Delete(N NetReq) int {
	if !N.IsAdmin() {
		return N.Error("Unauthorized", 403)
	}
	if util.DeleteUser(N.DB, N.Url[0]) {
		return N.OK()
	} 
	return N.Error("Failed", 500)
}



func (G Users) Patch(N NetReq) int {
	return N.NotFound()
}

func (G Users) Head(N NetReq) int {
	return N.NotFound()
}

func (G Users) ResourceName() string {
	return "user"
}
