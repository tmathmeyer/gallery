package web

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"encoding/json"
	"path/filepath"
	"strings"
	"io"
	"os/exec"


	"../database/util"
	"../database/generated"
)

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

		path := GetPathParts(r, "/api/v0/image/", 1)
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

func testImageType(filepath string) int {
	cmd := exec.Command("./exif/bin/photosphere", filepath)
	err := cmd.Run()
	if err != nil {
		return 0 // Not a panoramic
	} else {
		return 1 // A panoramic / photosphere
	}
}


func apiImageUploadHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	path := GetPathParts(r, "/api/v0/upload/", 2)
	if len(path) != 2 {
		http.NotFound(w, r)
		return
	}

	galleryPath := path[1]

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
	photo.Type = testImageType(writeToPath)
	photo.Name = filename+fileExtension
	photo.Description = filename+fileExtension
	photo.Gallery = galleryPath

	generated.InsertPhotoTable(db, photo)
	http.Error(w, filename+fileExtension, 200)
	return
}

func apiGpxUploadHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 15)
	path := GetPathParts(r, "/api/v0/upload/", 2)
	if len(path) != 2 {
		http.NotFound(w, r)
		return
	}

	galleryPath := path[1]

	imageStore := util.GetMetadataValue(db, "dataStore")

	file, _, err := r.FormFile("gpx")
	if err != nil {
		http.Error(w, "Upload failed", 500)
		fmt.Println(err)
		return
	}
	defer file.Close()

	writeToPath := fmt.Sprintf("%s/%s/route.gpx", imageStore, galleryPath)

	f, err := os.OpenFile(writeToPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		http.Error(w, "Upload failed", 500)
		return
	}
	defer f.Close()
	io.Copy(f, file)

	generated.UpdateGalleryTable(db, "hasgpx", 1, map[string]interface{}{
		"Path": galleryPath,
	})

	http.Error(w, "OK", 200)
}

func apiUploadHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c,e:=r.Cookie("admin"); e!=nil || c.Value!="admin" {
			http.Error(w, "Unauthorized", 403)
			return
		}

		path := GetPathParts(r, "/api/v0/upload/", 2)
		if len(path) != 2 {
			http.NotFound(w, r)
			return
		}

		switch path[0] {
		case "image":
			apiImageUploadHandler(db, w, r)
			return
		case "gpx":
			apiGpxUploadHandler(db, w, r)
			return
		}
		http.NotFound(w, r)
	})
}

func ApiHandler(db *sql.DB) http.Handler {
	types := map[string]http.Handler{
		"image": apiImageHandler(db),
		"upload": apiUploadHandler(db),
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := GetPathParts(r, "/api/v0/", 3)
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