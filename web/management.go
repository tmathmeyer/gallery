package web

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"../database/util"
	"../database/generated"
	"log"
	"os"
	"encoding/json"
	"path/filepath"
	"strings"
	"io"
)


// Creates a gallery
func apiGalleryHandlerPost(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	gn := r.Form["galleryname"]
	if len(gn) != 1 {
		http.Error(w, "missing galleryname", 400)
		return
	}

	galleryname := gn[0]

	dataStore := util.GetMetadataValue(db, "dataStore")
	imageStore := util.GetMetadataValue(db, "imageStore")

	if imageStore == "" || dataStore == "" {
		log.Fatal("Can't operate without data and image locations")
		return
	}

	localpath := util.MakeFriendlyPath(galleryname)
	for FExists(imageStore + "/" + localpath) {
		localpath = localpath + "0"
	}

	err := os.Mkdir(imageStore + "/" + localpath, os.ModePerm)
	if err != nil {
		http.Error(w, "Failed to make image store: "+localpath, 500)
	}
	err = os.Mkdir(dataStore + "/" + localpath, os.ModePerm)
	if err != nil {
		http.Error(w, "Failed to make data store: "+localpath, 500)
	}

	err = os.Symlink("../../static/placeholder.png", imageStore + "/" + localpath + "/480placeholder.png")
	if err != nil {
		http.NotFound(w, r)
		return
	}


	var gallery generated.Gallery
	gallery.Name = galleryname
	gallery.Path = localpath
	gallery.Splash = "placeholder.png"
	gallery.Lat = 0
	gallery.Lon = 0
	generated.InsertGalleryTable(db, gallery)
	
	http.Error(w, "OK", 200)
}

func apiGalleryHandlerPut(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	path := GetPathParts(r, "/api/gallery/", 1)
	if len(path) != 1 {
		http.NotFound(w, r)
		return
	}

	galleryPath := path[0]
	r.ParseForm()
	props := []string{"name", "splash", "lat", "lon"}
	for _, prop := range props {
		gn := r.Form[prop]
		if len(gn) == 1 {
			err := generated.UpdateGalleryTable(db, prop, gn[0], map[string]interface{}{
				"Path": galleryPath,
			})
			if err != nil {
				http.Error(w, "Failed to change property " + prop, 500)
				return
			}
		}
	}
	http.Error(w, "OK", 200)

}

func GalleryManagementHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t, _ := template.ParseFiles("templates/manage.html")

		galleries, err := generated.QueryGalleryTable(db, map[string]interface{}{})
		if err != nil {
			http.NotFound(w, r)
			return
		}

		title := util.GetMetadataValue(db, "siteName")
		apiKey := util.GetMetadataValue(db, "gmapsApiKey")

		var index = IndexModel{
			Title:     title,
			APIKey:    apiKey,
			Galleries: galleries}
		t.Execute(w, index)
	})
}

func apiGalleryHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c,e:=r.Cookie("admin"); e!=nil || c.Value!="admin" {
			http.Error(w, "Unauthorized", 403)
			return
		}

		switch r.Method {
		case "POST":
			apiGalleryHandlerPost(db, w, r)
			return

		case "PUT":
			apiGalleryHandlerPut(db, w, r)
			return
		}
		http.NotFound(w, r)
	})
}

func apiImageHandlerGet(db *sql.DB, w http.ResponseWriter, r *http.Request, gallery string) {
	photos, err := generated.QueryPhotoTable(db, map[string]interface{}{
		"Gallery": gallery,
	})
	if err != nil {
		http.Error(w, "Failed to lookup photos", 500)
		return
	}
	jData, err := json.Marshal(photos)
	if err != nil {
		panic(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}

func apiImageHandlerPut(db *sql.DB, w http.ResponseWriter, r *http.Request, gallery string) {
	r.ParseForm()
	imgs := r.Form["image"]
	descrs := r.Form["description"]

	if len(imgs) != 1 || len(descrs) != 1 {
		http.Error(w, "missing image or description", 400)
		return
	}

	err := generated.UpdatePhotoTable(db, "Description", descrs[0], map[string]interface{}{
		"Gallery": gallery,
		"Name": imgs[0],
	});

	if err != nil {
		http.Error(w, "Failed to set description", 500)
		return
	}

	http.Error(w, "OK", 200)
}

func apiImageHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c,e:=r.Cookie("admin"); e!=nil || c.Value!="admin" {
			http.Error(w, "Unauthorized", 403)
			return
		}

		path := GetPathParts(r, "/api/image/", 1)
		if len(path) == 0 {
			http.NotFound(w, r)
			return
		}
		galleries, err := generated.QueryGalleryTable(db, map[string]interface{}{
			"Path": path[0],
		})
		if err != nil || len(galleries) != 1 {
			http.Error(w, "Missing gallery", 500)
			return
		}


		switch r.Method {
		case "GET":
			apiImageHandlerGet(db, w, r, path[0])
			return

		case "PUT":
			apiImageHandlerPut(db, w, r, path[0])
			return
		}
		http.NotFound(w, r)
	})
}


func apiUserManagement(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := r.Cookie("username")
		if err != nil {
			http.Error(w, "Invalid Request", 400)
			return
		}

		r.ParseMultipartForm(32 << 20)
		switch r.Method {
		case "PUT":
			uidl := r.Form["id"]
			if len(uidl) == 1 && util.ChangePasswordForId(db, uidl[0], r.Form["password"][0]) {
				http.Error(w, "OK", 200)
			} else if len(uidl) != 1 && util.ChangePassword(db, user.Value, r.Form["password"][0]) {
				http.Error(w, "OK", 200)
			} else {
				http.Error(w, "Failed", 500)
			}
			return
		case "POST":
			util.AddUser(db, r.Form["username"][0], r.Form["password"][0])
			http.Error(w, "OK", 200)
			return
		case "DELETE":
			username := GetPathParts(r, "/api/user/", 1)[0]
			if util.DeleteUser(db, username) {
				http.Error(w, "OK", 200)
			} else {
				http.Error(w, "Failed", 500)
			}
			return
		case "GET":
			if c,e:=r.Cookie("admin"); e!=nil || c.Value!="admin" {
				http.Error(w, "Unauthorized", 403)
				return
			}
			users, err := generated.QueryUserTable(db, map[string]interface{}{})
			for i := range users {
				users[i].Passhash = "redacted"
			}
			jData, err := json.Marshal(users)
			if err != nil {
				panic(err)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(jData)
			return
		}

		http.Error(w, "Invalid Request", 400)
	})
}



func apiImageUploadHandlerPost(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	path := GetPathParts(r, "/api/upload/", 1)
	if len(path) != 1 {
		http.NotFound(w, r)
		return
	}

	galleryPath := path[0]

	imageStore := util.GetMetadataValue(db, "imageStore")

	file, handler, err := r.FormFile("newimage")
	if err != nil {
		http.Error(w, "Upload failed", 500)
		return
	}
	defer file.Close()

	fileExtension := filepath.Ext(handler.Filename)
	filename := strings.TrimSuffix(handler.Filename, fileExtension)

	writeToPath := fmt.Sprintf("%s/%s/%s%s", imageStore, galleryPath, filename, fileExtension)

	for FExists(writeToPath) {
		filename = filename + "0"
		writeToPath = fmt.Sprintf("%s/%s/%s%s", imageStore, galleryPath, filename, fileExtension)
	}

	f, err := os.OpenFile(writeToPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		http.Error(w, "Upload failed", 500)
		return
	}
	defer f.Close()
	io.Copy(f, file)

	var photo generated.Photo
	photo.Type = 0
	photo.Name = filename+fileExtension
	photo.Description = filename+fileExtension
	photo.Gallery = galleryPath

	generated.InsertPhotoTable(db, photo)
	http.Error(w, filename+fileExtension, 200)
	return
}

func apiImageUploadHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c,e:=r.Cookie("admin"); e!=nil || c.Value!="admin" {
			http.Error(w, "Unauthorized", 403)
			return
		}

		switch r.Method {
		case "POST":
			apiImageUploadHandlerPost(db, w, r)
			return
		}
		http.NotFound(w, r)
	})
}

func ApiHandler(db *sql.DB) http.Handler {
	types := map[string]http.Handler{
		"gallery": apiGalleryHandler(db),
		"image": apiImageHandler(db),
		"upload": apiImageUploadHandler(db),
		"user": apiUserManagement(db),
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := GetPathParts(r, "/api/", 3)
		if len(path) == 0 {
			http.NotFound(w, r)
			return
		}
		if handler, ok := types[path[0]]; ok {
			handler.ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
		return
	})
}