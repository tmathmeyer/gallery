package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const serverPort string = ":8081"

type Image struct {
	name string
	Desc string
}

type GalleryDetailModel struct {
	GalleryName string
	GalleryID   string
	Images      []Photo
	Title       string
	GpxPresent  bool
	APIKey      string
}

type CustomImageViewModel struct {
	Title            string
	GalleryName      string
	GalleryPath      string
	ImagePath        string
	ImageDescription string
	ImageURL         string
}

type IndexModel struct {
	Title     string
	APIKey    string
	Galleries []Gallery
}

// Test if file exists
func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func getPathParts(r *http.Request, str string, splitct int) []string {
	restPath := r.URL.Path[len(str):]
	return strings.SplitN(restPath, "/", splitct)
}

// GET /i/:size/:gallery/:photo
func imageHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url_data := getPathParts(r, "/i/", 3)
		if len(url_data) != 3 {
			fmt.Printf("malformed URL\n")
			http.NotFound(w, r)
			return
		}

		// #security
		for _, dirname := range url_data[:len(url_data)-1] {
			if strings.HasPrefix(dirname, ".") {
				http.NotFound(w, r)
				fmt.Printf("%s : [%s]\n", "Cannot start with a '.'", dirname)
				return
			}
		}

		var prefix = ""
		var size = "450"
		switch url_data[0] {
		case "F":
			prefix = ""
		case "L":
			prefix = "lg_"
			size = "2560"
		case "M":
			prefix = "md_"
			size = "1920"
		case "S":
			prefix = "sm_"
			size = "1280"
		case "T":
			prefix = "tn_"
			size = "640"
		default:
			prefix = "tn_"
		}

		galleryData := getMetadataValue(db, "galleryData")

		image_path := galleryData + url_data[1] + "/" + prefix + url_data[2]

		if exists(image_path) {
			http.ServeFile(w, r, image_path)
			return
		}

		if url_data[0] != "F" {
			original_path := galleryData + url_data[1] + "/" + url_data[2]
			pattern := prefix + "%s" + filepath.Ext(original_path)
			resize_and_serve(w, r, original_path, image_path, pattern, size)
		}
	})
}

func resize_and_serve(w http.ResponseWriter, r *http.Request, original string, to_create string, pattern string, size string) {
	cmd := exec.Command("vipsthumbnail", "-s", size, "-o", pattern, original)
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		fmt.Printf("vipsthumbnail -s %s -o %s %s\n", size, pattern, original)
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, to_create)
}

// GET /g/:gallery
func galleryDetailhandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		galleryID := getPathParts(r, "/g/", 1)[0]

		var gallery Gallery
		if err := getGallery(db, galleryID, &gallery); err != nil {
			fmt.Printf("%s\n", err)
			http.NotFound(w, r)
			return
		}

		photos, err := getPhotosByGallery(db, galleryID)
		if err != nil {
			fmt.Printf("%s\n", err)
			http.NotFound(w, r)
			return
		}

		title := getMetadataValue(db, "title")
		apiKey := getMetadataValue(db, "mapapikey")
		galleryData := getMetadataValue(db, "galleryData")

		var gpxPresent = false
		if _, err := os.Stat(galleryData + galleryID + "/route.gpx"); err == nil {
			gpxPresent = true
		}

		var Detail = GalleryDetailModel{
			GalleryName: gallery.Name,
			GalleryID:   gallery.Path,
			Images:      photos,
			Title:       title,
			GpxPresent:  gpxPresent,
			APIKey:      apiKey,
		}

		t, _ := template.ParseFiles("templates/detail.html")
		t.Execute(w, Detail)
		return

	})
}

// GET /v/:type/:gallery/:image
func customViewHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		viewdata := getPathParts(r, "/v/", 3)
		galleryID := viewdata[1]
		photoName := viewdata[2]

		var gallery Gallery
		if err := getGallery(db, galleryID, &gallery); err != nil {
			fmt.Printf("%s\n", err)
			http.NotFound(w, r)
			return
		}

		var photo Photo
		err := getPhotosByGalleryAndName(db, galleryID, photoName, &photo)
		if err != nil {
			fmt.Printf("%s\n", err)
			http.NotFound(w, r)
			return
		}

		title := getMetadataValue(db, "title")

		var Detail = CustomImageViewModel{
			Title:            title,
			GalleryName:      gallery.Name,
			GalleryPath:      galleryID,
			ImagePath:        photoName,
			ImageDescription: photo.Descr,
			ImageURL:         "/i/F/" + galleryID + "/" + photoName,
		}

		t, _ := template.ParseFiles("templates/image.html")
		t.Execute(w, Detail)
	})
}

func indexHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t, _ := template.ParseFiles("templates/index.html")

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

func setupPublicHandlers(db *sql.DB) {
	http.Handle("/", indexHandler(db))
	http.Handle("/g/", galleryDetailhandler(db))
	http.Handle("/v/", customViewHandler(db))
	http.Handle("/i/", imageHandler(db))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
}

func main() {
	db, err := open_database("./live.sqlite")
	if err != nil {
		log.Fatal(err)
	}

	setupPublicHandlers(db)
	setupPrivateHandlers(db)

	fmt.Printf("Starting gallery server on port %s\n", serverPort)
	http.ListenAndServe(serverPort, nil)
}
