package main

import (
	"os"
	"fmt"
	"strings"
	"net/http"
	"html/template"
	"github.com/BurntSushi/toml"
)

type Hike struct {
	HikeName string
	HikeID string
	HikeLat string
	HikeLng string
}

type Config struct {
	Title string
	ApiKey string
	Hikes []Hike
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("index.template.html")
	var config Config
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
	  	fmt.Printf("%s", err)
	  	http.NotFound(w, r)
	}
	t.Execute(w, config)
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return true, err
}

func hikeDataHandler(w http.ResponseWriter, r *http.Request) {
	restPath := r.URL.Path[len("/hikedata/"):]
	partsPath := strings.Split(restPath, "/")
	if len(partsPath) < 2 {
		http.NotFound(w, r)
		fmt.Printf("%s : [%s]", "path has too few parts!", restPath)
		return
	}
	if strings.HasPrefix(partsPath[0], ".") {
		http.NotFound(w, r)
		fmt.Printf("%s : [%s]", "Cannot start with a '.'", partsPath[0])
		return
	}
	if strings.HasPrefix(partsPath[1], ".") {
		http.NotFound(w, r)
		fmt.Printf("%s : [%s]", "Cannot start with a '.'", partsPath[1])
		return
	}
	fse,_ := exists("./hikedata/" + restPath)
	if !fse {
		http.NotFound(w, r)
		fmt.Printf("%s : [%s]", "path does not exist", restPath)
		return
	}

	http.ServeFile(w, r, "./hikedata/" + restPath)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/hikedata/", hikeDataHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.ListenAndServe(":8081", nil)
}