package main

import (
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

// Hike stores hike information for displaying on index page
type Hike struct {
	HikeName string
	HikeID   string
	HikeLat  string
	HikeLng  string
}

// Config stores server configuration
type Config struct {
	Title  string
	APIKey string
	Hikes  []Hike
}

// Image stores locator and description
type Image struct {
	URI  string
	Desc string
}

// HikeDetail stores hike information for displaying on contents page
type HikeDetail struct {
	HikeName string
	HikeID   string
	Images   []Image
	Panos    []Image
}

func hikeDetailhandler(w http.ResponseWriter, req *http.Request) {
	var config Config
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		fmt.Printf("%s\n", err)
		http.NotFound(w, req)
	}
	hikeID := req.URL.Path[len("/gallerydetail/"):]
	for _, Hike := range config.Hikes {
		if Hike.HikeID == hikeID {
			img, _ := ioutil.ReadDir(galleryDataDir + hikeID + "/img")
			pan, _ := ioutil.ReadDir(galleryDataDir + hikeID + "/pan")

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

			var Detail = HikeDetail{
				HikeName: Hike.HikeName,
				HikeID:   Hike.HikeID,
				Panos:    Panos,
				Images:   Images}

			t, _ := template.ParseFiles("detail.template.html")
			t.Execute(w, Detail)
			return
		}
	}
	http.NotFound(w, req)
}

func imageHandler(w http.ResponseWriter, req *http.Request) {
	// Path expected: /img/[[size]]/[[HikeId]]/[[ImageName]]
	typepath := strings.SplitN(req.URL.Path[1:], "/", 4)
	if len(typepath) != 4 {
		http.NotFound(w, req)
		return
	}

	for _, dirname := range typepath[:len(typepath)-1] {
		if strings.HasPrefix(dirname, ".") {
			http.NotFound(w, req)
			fmt.Printf("%s : [%s]\n", "Cannot start with a '.'", dirname)
			return
		}
	}

	// ie img/ or pan/
	imgDir := typepath[0] + "/"
	imgSize := typepath[1]
	hikeID := typepath[2] + "/"
	imgName := typepath[3]

	imgPath := galleryDataDir + hikeID + imgDir + imgName
	if imgSize == "full" {
		http.ServeFile(w, req, imgPath)
		return
	}

	if imgSize == "small" {
		// Jank code to replace file extensions
		thmPath := galleryDataDir + hikeID + imgDir + thumbPrefix + strings.TrimSuffix(imgName, filepath.Ext(imgName)) + ".jpg"

		if !exists(thmPath) {
			// vipsthumbnail will only produce .jpg images
			fmt.Printf("Creating thumbnail for Image: %s\n", thmPath)
			cmd := exec.Command("vipsthumbnail", "-s", "450", imgPath)
			err := cmd.Run()
			if err != nil {
				fmt.Println(err)
				http.NotFound(w, req)
				return
			}
		}
		http.ServeFile(w, req, thmPath)
		return
	}

	fmt.Printf("req : [%s]\n", hikeID)
	http.NotFound(w, req)
}

func indexHandler(w http.ResponseWriter, req *http.Request) {
	t, _ := template.ParseFiles("index.template.html")
	var config Config
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		fmt.Printf("%s\n", err)
		http.NotFound(w, req)
	}
	// Create directories if they do not exist
	mkdir(galleryDataDir)
	for _, hike := range config.Hikes {
		mkdir(galleryDataDir + hike.HikeID)
		mkdir(galleryDataDir + hike.HikeID + "/img")
		mkdir(galleryDataDir + hike.HikeID + "/pan")
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

func mkdir(path string) {
	if exists(path) {
		return
	}
	err := os.Mkdir(path, os.ModePerm)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}

func hikeDataHandler(w http.ResponseWriter, r *http.Request) {
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

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/gallerydetail/", hikeDetailhandler)
	http.HandleFunc("/gallerydata/", hikeDataHandler)
	http.HandleFunc("/img/", imageHandler)
	http.HandleFunc("/pan/", imageHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	fmt.Printf("Starting gallery server on port %s\n", serverPort)
	http.ListenAndServe(serverPort, nil)
}
