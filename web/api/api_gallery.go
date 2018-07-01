package api

import (
	"net/http"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"../../web"
	"../../database/generated"
	"../../database/util"
)

type Gallery struct {
	DB *sql.DB
}

func (G Gallery) GetDatabase() *sql.DB {
	return G.DB
}

func (G Gallery) Get(w http.ResponseWriter, r *http.Request, url []string) {
	if len(url) == 0 {
		galleries, err := generated.QueryGalleryTable(G.DB, map[string]interface{}{})
		if err != nil {
			http.NotFound(w, r);
			return
		}
		data, err := json.Marshal(galleries)
		if err != nil {
			http.NotFound(w, r);
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
		return
	}
	if len(url) == 1 {
		galleries, err := generated.QueryGalleryTable(G.DB, map[string]interface{}{
			"Path": url[0],
		})
		if err != nil {
			http.NotFound(w, r);
			return
		}
		if len(galleries) != 1 {
			http.NotFound(w, r);
			return
		}
		data, err := json.Marshal(galleries[0])
		if err != nil {
			http.NotFound(w, r);
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
		return
	}
	http.NotFound(w, r)
}

func (G Gallery) Post(w http.ResponseWriter, r *http.Request, url []string) {
	r.ParseForm()
	gn := r.Form["galleryname"]
	if len(gn) != 1 {
		http.Error(w, "missing galleryname", 400)
		return
	}

	galleryname := gn[0]

	dataStore := util.GetMetadataValue(G.DB, "dataStore")
	imageStore := util.GetMetadataValue(G.DB, "imageStore")

	if imageStore == "" || dataStore == "" {
		log.Fatal("Can't operate without data and image locations")
		return
	}

	localpath := util.MakeFriendlyPath(galleryname)
	for web.FExists(imageStore + "/" + localpath) {
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
	generated.InsertGalleryTable(G.DB, gallery)
	
	http.Error(w, "OK", 200)
}

func (G Gallery) Put(w http.ResponseWriter, r *http.Request, url []string) {
	if len(url) != 1 {
		http.NotFound(w, r)
		return
	}

	r.ParseForm()
	props := []string{"name", "splash", "lat", "lon", "hasgpx"}
	for _, prop := range props {
		gn := r.Form[prop]
		if len(gn) == 1 {
			err := generated.UpdateGalleryTable(G.DB, prop, gn[0], map[string]interface{}{
				"Path": url[0],
			})
			if err != nil {
				http.Error(w, "Failed to change property " + prop, 500)
				return
			}
		}
	}
	http.Error(w, "OK", 200)
}

func (G Gallery) Delete(w http.ResponseWriter, r *http.Request, url []string) {
	http.NotFound(w, r)
}

func (G Gallery) Patch(w http.ResponseWriter, r *http.Request, url []string) {
	http.NotFound(w, r)
}

func (G Gallery) Head(w http.ResponseWriter, r *http.Request, url []string) {
	http.NotFound(w, r)
}

func (G Gallery) ResourceName() string {
	return "gallery"
}
