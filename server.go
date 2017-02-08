package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type Hike struct {
	HikeName string
	HikeID   string
	HikeLat  string
	HikeLng  string
}

type Config struct {
	Title  string
	ApiKey string
	Hikes  []Hike
}

type Image struct {
	Uri  string
	Desc string
}

type Pano struct {
	Uri string
	HikeID string
}

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
					imgct += 1
				}
			}
			for _, f := range pan {
				if !strings.HasPrefix(f.Name(), "tn_") {
					panct += 1
				}
			}

			var Images = make([]Image, imgct)
			for i, f := range img {
				if !strings.HasPrefix(f.Name(), "tn_") {
					Images[i] = Image{
						Uri:  f.Name(),
						Desc: f.Name()}
				}
			}

			var Panos = make([]Image, panct)
			for i, f := range pan {
				if !strings.HasPrefix(f.Name(), "tn_") {
					Panos[i] = Image{
						Uri:  f.Name(),
						Desc: f.Name()}
				}
			}

			var Detail = HikeDetail{
				HikeName: Hike.HikeName,
				HikeID:   Hike.HikeID,
				Panos:    Panos,
				Images:   Images}

			t, _ := template.ParseFiles("templates/detail.html")
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

		http.ServeFile(w, r, thm_path)
		return
	}

	fmt.Printf("req : [%s]\n", typepath[1])
	http.NotFound(w, r)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/index.html")
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

func panoramicHandler(w http.ResponseWriter, r *http.Request) {
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

	if typepath[0] == "raw" {
		img_path := "./gallerydata/" + typepath[1] + "/pan/" + typepath[2]
		http.ServeFile(w, r, img_path)
		return
	}

	if typepath[0] == "full" {
		t, _ := template.ParseFiles("templates/pano_view.html")
        var img = Pano{Uri:  typepath[2],
			 	       HikeID: typepath[1]}
		t.Execute(w, img)
		return
	}

	if typepath[0] == "small" {
		img_path := "./gallerydata/" + typepath[1] + "/pan/" + typepath[2]
		thm_path := "./gallerydata/" + typepath[1] + "/pan/tn_" + typepath[2]

		if !exists(thm_path) {
			cmd := exec.Command("vipsthumbnail", "-s", "450", img_path)
			err := cmd.Run()
			if err != nil {
				fmt.Println(err)
				http.NotFound(w, r)
				return
			}
		}

		http.ServeFile(w, r, thm_path)
		return
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
	http.HandleFunc("/pan/", panoramicHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.ListenAndServe(":8081", nil)
}
