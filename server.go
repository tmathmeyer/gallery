package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/BurntSushi/toml"
)

const serverPort string = ":8081"

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

func hikeDetailhandler(w http.ResponseWriter, r *http.Request) {
	var config Config
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		fmt.Printf("%s\n", err)
		http.NotFound(w, r)
	}
	hikeID := r.URL.Path[len("/gallerydetail/"):]
	for _, Hike := range config.Hikes {
		if Hike.HikeID == hikeID {
			img, _ := ioutil.ReadDir("./gallerydata/" + hikeID + "/img")
			pan, _ := ioutil.ReadDir("./gallerydata/" + hikeID + "/pan")

			var imgct = 0
			var panct = 0
			for _, f := range img {
				if !strings.HasPrefix(f.Name(), "tn_") {
					imgct++
				}
			}
			for _, f := range pan {
				if !strings.HasPrefix(f.Name(), "tn_") {
					panct++
				}
			}

			var Images = make([]Image, imgct)
			for i, f := range img {
				if !strings.HasPrefix(f.Name(), "tn_") {
					Images[i] = Image{
						URI:  f.Name(),
						Desc: f.Name()}
				}
			}

			var Panos = make([]Image, panct)
			for i, f := range pan {
				if !strings.HasPrefix(f.Name(), "tn_") {
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
	http.NotFound(w, r)
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	typepath := strings.SplitN(r.URL.Path[len("/img/"):], "/", 3)
	if len(typepath) != 3 {
		http.NotFound(w, r)
		return
	}

	if strings.HasPrefix(typepath[0], ".") {
		http.NotFound(w, r)
		fmt.Printf("%s : [%s]\n", "Cannot start with a '.'", typepath[0])
		return
	}
	if strings.HasPrefix(typepath[1], ".") {
		http.NotFound(w, r)
		fmt.Printf("%s : [%s]\n", "Cannot start with a '.'", typepath[1])
		return
	}

	if typepath[0] == "full" {
		img_path := "./gallerydata/" + typepath[1] + "/img/" + typepath[2]
		http.ServeFile(w, r, img_path)
		return
	}

	if typepath[0] == "small" {
		img_path := "./gallerydata/" + typepath[1] + "/img/" + typepath[2]
		thm_path := "./gallerydata/" + typepath[1] + "/img/tn_" + typepath[2]

		if !exists(thm_path) {
			cmd := exec.Command("./vipsthumbnail", "-s", "450", img_path)
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

	fmt.Printf("req : [%s]\n", typepath[1])
	http.NotFound(w, r)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("index.template.html")
	var config Config
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		fmt.Printf("%s\n", err)
		http.NotFound(w, r)
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
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	fmt.Printf("Starting gallery server on port %s\n", serverPort)
	http.ListenAndServe(serverPort, nil)
}
