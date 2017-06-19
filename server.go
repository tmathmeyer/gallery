package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const serverPort string = ":8081"
const galleryDataDir string = "./gallerydata/"
const thumbPrefix string = "tn_"
const copyrightHolder = "your.email@cock.li"

type Gallery struct {
	GalleryName string
	GalleryID   string
	GalleryLat  string
	GalleryLng  string
}

type Config struct {
	Title     string
	APIKey    string
	Galleries []Gallery
}

type Image struct {
	URI  string
	Desc string
}

type Pano struct {
	URI             string
	GalleryID       string
	CopyrightHolder string
}

type GalleryDetail struct {
	GalleryName string
	GalleryID   string
	Images      []Image
	Panos       []Image
	Title       string
	GpxPresent  bool
	APIKey      string
}

type ImageTemplate struct {
	Title       string
	GalleryID   string
	GalleryName string
	URI         string
	Image       string
}

func getConfig(config *Config) error {
	if _, err := toml.DecodeFile("config.toml", config); err != nil {
		fmt.Printf("%s\n", err)
		return errors.New("Cannot get configuration!")
	}
	return nil
}

func getPathParts(r *http.Request, str string) []string {
	restPath := r.URL.Path[len(str):]
	return strings.Split(restPath, "/")
}

func getGallery(galleryID string, gallery *Gallery) error {
	var config Config
	if err := getConfig(&config); err != nil {
		return errors.New("Cannot get gallery; configuration missing")
	}
	for _, Gallery := range config.Galleries {
		if Gallery.GalleryID == galleryID {
			*gallery = Gallery
			return nil
		} else {
			fmt.Printf("(%s) != (%s)\n", galleryID, Gallery.GalleryID)
		}
	}
	fmt.Printf("Failed to lookup: (%s)\n", galleryID)
	return errors.New("Cannot get gallery; not present in configuration")
}

func galleryDetailhandler(w http.ResponseWriter, r *http.Request) {
	var config Config
	var gallery Gallery
	if err := getConfig(&config); err != nil {
		fmt.Printf("%s\n", err)
		http.NotFound(w, r)
		return
	}

	galleryID := getPathParts(r, "/gallerydetail/")[0]

	if err := getGallery(galleryID, &gallery); err != nil {
		fmt.Printf("%s\n", err)
		http.NotFound(w, r)
		return
	}

	img, _ := ioutil.ReadDir(galleryDataDir + galleryID + "/img")
	pan, _ := ioutil.ReadDir(galleryDataDir + galleryID + "/pan")

	var imgct = 0
	var panct = 0
	for _, f := range img {
		if !strings.HasPrefix(f.Name(), thumbPrefix) {
			imgct++
		}
	}
	for _, f := range pan {
		if !strings.HasPrefix(f.Name(), thumbPrefix) {
			panct++
		}
	}

	var Images = make([]Image, imgct)
	for i, f := range img {
		if !strings.HasPrefix(f.Name(), thumbPrefix) {
			Images[i] = Image{
				URI:  f.Name(),
				Desc: f.Name()}
		}
	}

	var Panos = make([]Image, panct)
	for i, f := range pan {
		if !strings.HasPrefix(f.Name(), thumbPrefix) {
			Panos[i] = Image{
				URI:  f.Name(),
				Desc: f.Name()}
		}
	}

	var gpxPresent = false
	if _, err := os.Stat(galleryDataDir + galleryID + "/route.gpx"); err == nil {
		gpxPresent = true
	}

	var Detail = GalleryDetail{
		GalleryName: gallery.GalleryName,
		GalleryID:   gallery.GalleryID,
		Panos:       Panos,
		Images:      Images,
		Title:       config.Title,
		GpxPresent:  gpxPresent,
		APIKey:      config.APIKey,
	}

	t, _ := template.ParseFiles("templates/detail.html")
	t.Execute(w, Detail)
	return
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	typepath := strings.SplitN(r.URL.Path[1:], "/", 4)
	if len(typepath) != 4 {
		http.NotFound(w, r)
		return
	}

	for _, dirname := range typepath[:len(typepath)-1] {
		if strings.HasPrefix(dirname, ".") {
			http.NotFound(w, r)
			fmt.Printf("%s : [%s]\n", "Cannot start with a '.'", dirname)
			return
		}
	}

	imgSize := typepath[1]
	galleryID := typepath[2]
	imgName := typepath[3]

	imgPath := galleryDataDir + galleryID + "/img/" + imgName
	if imgSize == "full" {
		http.ServeFile(w, r, imgPath)
		return
	}

	if imgSize == "small" {
		thmPath := galleryDataDir + galleryID + "/img/" + thumbPrefix + strings.TrimSuffix(imgName, filepath.Ext(imgName)) + ".jpg"

		if !exists(thmPath) {
			cmd := exec.Command("vipsthumbnail", "-s", "450", imgPath)
			err := cmd.Run()
			if err != nil {
				fmt.Println(err)
				http.NotFound(w, r)
				return
			}
		}
		http.ServeFile(w, r, thmPath)
		return
	}

	fmt.Printf("req : [%s]\n", galleryID)
	http.NotFound(w, r)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/index.html")
	var config Config
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		fmt.Printf("%s\n", err)
		http.NotFound(w, r)
	}
	for _, gallery := range config.Galleries {
		mkdir(galleryDataDir + gallery.GalleryID)
		mkdir(galleryDataDir + gallery.GalleryID + "/img")
		mkdir(galleryDataDir + gallery.GalleryID + "/pan")
	}
	t.Execute(w, config)
}

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

func panoramicHandler(w http.ResponseWriter, r *http.Request) {
	typepath := strings.SplitN(r.URL.Path[1:], "/", 4)
	if len(typepath) != 4 {
		http.NotFound(w, r)
		return
	}

	for _, dirname := range typepath[:len(typepath)-1] {
		if strings.HasPrefix(dirname, ".") {
			http.NotFound(w, r)
			fmt.Printf("%s : [%s]\n", "Cannot start with a '.'", dirname)
			return
		}
	}

	imgSize := typepath[1]
	galleryID := typepath[2]
	imgName := typepath[3]
	imgPath := "./gallerydata/" + galleryID + "/pan/" + imgName
	if imgSize == "raw" {
		http.ServeFile(w, r, imgPath)
		return
	}
	if imgSize == "full" {
		t, _ := template.ParseFiles("templates/pano_view.html")
		var img = Pano{URI: imgName, GalleryID: galleryID, CopyrightHolder: copyrightHolder}
		t.Execute(w, img)
		return
	}
	if imgSize == "small" {
		thmPath := galleryDataDir + galleryID + "/pan/" + thumbPrefix + strings.TrimSuffix(imgName, filepath.Ext(imgName)) + ".jpg"
		if !exists(thmPath) {
			cmd := exec.Command("vipsthumbnail", "-s", "450", imgPath)
			err := cmd.Run()
			if err != nil {
				fmt.Println(err)
				http.NotFound(w, r)
				return
			}
		}
		http.ServeFile(w, r, thmPath)
		return
	}
}

func mkdir(path string) {
	if exists(path) {
		return
	}
	err := os.Mkdir(path, os.ModePerm)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}

func galleryDataHandler(w http.ResponseWriter, r *http.Request) {
	restPath := r.URL.Path[len("/gallerydata/"):]
	partsPath := strings.Split(restPath, "/")
	if len(partsPath) < 2 {
		http.NotFound(w, r)
		fmt.Printf("%s : [%s]\n", "path has too few parts!", restPath)
		return
	}
	if strings.HasPrefix(partsPath[0], ".") {
		http.NotFound(w, r)
		fmt.Printf("%s : [%s]\n", "Cannot start with a '.'", partsPath[0])
		return
	}
	if strings.HasPrefix(partsPath[1], ".") {
		http.NotFound(w, r)
		fmt.Printf("%s : [%s]\n", "Cannot start with a '.'", partsPath[1])
		return
	}
	if !exists("./gallerydata/" + restPath) {
		http.NotFound(w, r)
		fmt.Printf("%s : [%s]\n", "path does not exist", restPath)
		return
	}

	http.ServeFile(w, r, "./gallerydata/"+restPath)
}

func galleryGpxHandler(w http.ResponseWriter, r *http.Request) {
	var config Config
	var gallery Gallery

	if err := getConfig(&config); err != nil {
		fmt.Printf("%s\n", err)
		http.NotFound(w, r)
		return
	}

	urlParts := getPathParts(r, "/gallerygpx/")
	galleryID := urlParts[0]

	if err := getGallery(galleryID, &gallery); err != nil {
		fmt.Printf("%s\n", err)
		http.NotFound(w, r)
		return
	}

	if len(urlParts) == 1 {
		var Detail = GalleryDetail{
			GalleryName: gallery.GalleryName,
			GalleryID:   gallery.GalleryID,
			Panos:       nil,
			Images:      nil,
			Title:       config.Title,
			GpxPresent:  false,
			APIKey:      config.APIKey,
		}

		t, _ := template.ParseFiles("templates/gpx.html")
		t.Execute(w, Detail)
		return
	}

	if len(urlParts) == 2 && urlParts[0] == "raw" {
		gpxPath := galleryDataDir + urlParts[1] + "/route.gpx"
		http.ServeFile(w, r, gpxPath)
		return
	}

	fmt.Printf("urlParts[0] isn't raw, its (%s)\n", urlParts[0])
	http.NotFound(w, r)
}

func galleryRawHandler(w http.ResponseWriter, r *http.Request) {
	var config Config
	var gallery Gallery

	if err := getConfig(&config); err != nil {
		fmt.Printf("%s\n", err)
		http.NotFound(w, r)
		return
	}

	urlParts := getPathParts(r, "/galleryraw/")
	galleryID := urlParts[0]

	if err := getGallery(galleryID, &gallery); err != nil {
		fmt.Printf("%s\n", err)
		http.NotFound(w, r)
		return
	}

	var Detail = ImageTemplate{
		Title:       config.Title,
		GalleryID:   gallery.GalleryID,
		GalleryName: gallery.GalleryName,
		URI:         "/img/full/" + gallery.GalleryID + "/" + urlParts[1],
		Image:       urlParts[1],
	}

	t, _ := template.ParseFiles("templates/image.html")
	t.Execute(w, Detail)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/gallerydetail/", galleryDetailhandler)
	http.HandleFunc("/gallerydata/", galleryDataHandler)
	http.HandleFunc("/gallerygpx/", galleryGpxHandler)
	http.HandleFunc("/galleryraw/", galleryRawHandler)
	http.HandleFunc("/img/", imageHandler)
	http.HandleFunc("/pan/", panoramicHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	fmt.Printf("Starting gallery server on port %s\n", serverPort)
	http.ListenAndServe(serverPort, nil)
}
