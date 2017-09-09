package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

func loginPageHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/login.html")
	})
}

func directToManagement() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/manage", 302)
	})
}

func loginRequestHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		user_a := r.Form["username"]
		pass_a := r.Form["password"]
		if len(user_a) != 1 || len(pass_a) != 1 {
			http.Error(w, "Unauthorized", 400)
			return
		}
		user := user_a[0]
		pass := pass_a[0]

		if is_admin(db, user, pass) {
			secret := getMetadataValue(db, "secret")
			token, err := get_authentication_token([]byte(secret), user)

			if err != nil {
				http.Error(w, "Unauthorized", 400)
				return
			}

			expiration := time.Now().Add(365 * 24 * time.Hour)
			cookie := http.Cookie{Name: "jwt", Value: token, Expires: expiration, Path: "/"}
			http.SetCookie(w, &cookie)
			directToManagement().ServeHTTP(w, r)
		} else {
			http.Error(w, "Unauthorized", 400)
		}
	})
}

func galleryManagementHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t, _ := template.ParseFiles("templates/manage.html")

		galleries, err := getGalleries(db)
		if err != nil {
			fmt.Printf("%s\n", err)
			http.NotFound(w, r)
			return
		}

		title := getMetadataValue(db, "title")
		apiKey := getMetadataValue(db, "apikey")

		var index = IndexModel{
			Title:     title,
			APIKey:    apiKey,
			Galleries: galleries}
		t.Execute(w, index)
	})
}

func apiHandler(db *sql.DB, types map[string]http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := getPathParts(r, "/a/", 3)
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

func removeSpaces(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}

func apiImageUploadHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			r.ParseMultipartForm(32 << 20)
			path := getPathParts(r, "/a/u/", 1)
			if len(path) == 0 {
				http.NotFound(w, r)
				return
			}

			galleryPath := path[0]
			prefix := getMetadataValue(db, "galleryData")

			file, handler, err := r.FormFile("newimage")
			if err != nil {
				http.Error(w, "Upload failed", 500)
				fmt.Println(err)
				return
			}

			defer file.Close()

			FileExtension := filepath.Ext(handler.Filename)
			Filename := strings.TrimSuffix(handler.Filename, FileExtension)

			writeToPath := prefix + galleryPath + "/" + Filename + FileExtension

			for file_exists(writeToPath) {
				Filename = Filename + "0"
				writeToPath = prefix + galleryPath + "/" + Filename + FileExtension
			}

			f, err := os.OpenFile(writeToPath, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				http.Error(w, "Upload failed", 500)
				fmt.Println(err)
				return
			}

			defer f.Close()
			io.Copy(f, file)

			add_photo(db, Filename+FileExtension, Filename+FileExtension, galleryPath)
			http.Error(w, Filename+FileExtension, 200)
			return
		}
	})
}

func apiImageHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := getPathParts(r, "/a/i/", 1)
		if len(path) == 0 {
			http.NotFound(w, r)
			return
		}
		galleryPath := path[0]

		switch r.Method {
		case "GET":
			photos, err := getPhotosByGallery(db, galleryPath)

			if err != nil {
				http.Error(w, "missing galleryname", 500)
				return
			}

			jData, err := json.Marshal(photos)
			if err != nil {
				panic(err)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(jData)
			return

		case "PUT":
			r.ParseForm()

			imgs := r.Form["image"]
			descrs := r.Form["description"]

			if len(imgs) != 1 || len(descrs) != 1 {
				http.Error(w, "missing image or description", 400)
				return
			}

			err := set_image_description_by_name_and_gallery(db, galleryPath, imgs[0], descrs[0])
			if err != nil {
				fmt.Printf("%s\n", err)
				http.Error(w, "DB", 500)
				return
			}

			http.Error(w, "OK", 200)
			return
		}
	})
}

func apiGalleryHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			r.ParseForm()
			gn := r.Form["galleryname"]
			if len(gn) != 1 {
				http.Error(w, "missing galleryname", 400)
				return
			}

			//make the directory
			//symlink the default splash
			//add_gallery(db *sql.DB, name string, path string, 90, 0, "placeholder.png")

			prefix := getMetadataValue(db, "galleryData")
			if prefix == "" {
				http.NotFound(w, r)
				return
			}

			localpath := removeSpaces(gn[0])
			path := prefix + "/" + localpath
			for file_exists(path) {
				path = path + "0"
				localpath = localpath + "0"
			}

			err := os.Mkdir(path, os.ModePerm)
			if err != nil {
				http.NotFound(w, r)
				return
			}

			err = os.Symlink("../../static/placeholder.png", path+"/tn_placeholder.png")
			if err != nil {
				http.NotFound(w, r)
				return
			}

			add_gallery(db, gn[0], localpath, 0, 0, "placeholder.png")

			fmt.Printf("Creating gallery: %s\n", gn[0])
			http.Error(w, "OK", 200)
			return

		case "PUT":
			path := getPathParts(r, "/a/g/", 1)
			if len(path) == 0 {
				fmt.Printf("path = %s\n", path)
				http.NotFound(w, r)
				return
			}
			gallPath := path[0]
			r.ParseForm()

			props := []string{"name", "splash", "lat", "lon"}
			for _, prop := range props {
				gn := r.Form[prop]
				if len(gn) == 1 {
					err := set_gallery_prop(db, gallPath, prop, gn[0])
					if err != nil {
						fmt.Printf("%s\n", err)
						http.Error(w, "DB", 500)
						return
					}
				}
			}

			http.Error(w, "OK", 200)
			return
		}
		http.NotFound(w, r)
	})
}

func setupPrivateHandlers(db *sql.DB) {
	http.Handle("/manage", verify_authentication_middleware(galleryManagementHandler(db), loginPageHandler(), db))
	http.Handle("/auth/handle", loginRequestHandler(db))

	apiTypes := map[string]http.Handler{
		"g": apiGalleryHandler(db),
		"i": apiImageHandler(db),
		"u": apiImageUploadHandler(db),
	}
	http.Handle("/a/", verify_authentication_middleware(apiHandler(db, apiTypes), nil, db))
}
