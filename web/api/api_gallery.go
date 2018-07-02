package api

import (
	"encoding/json"
	"os"
	"../../web"
	"../../database/generated"
	"../../database/util"
)

type Gallery struct {}

func (G Gallery) Get(N NetReq) int {
	N.W.Header().Set("Content-Type", "application/json")

	if len(N.Url) == 0 {
		galleries, err := generated.QueryGalleryTable(N.DB, map[string]interface{}{})
		if err != nil {
			return N.NotFound()
		}
		data, err := json.Marshal(galleries)
		if err != nil {
			return N.NotFound()
		}
		N.Write(data)
		return 200
	} else if len(N.Url) == 1 {
		galleries, err := generated.QueryGalleryTable(N.DB, map[string]interface{}{
			"Path": N.Url[0],
		})
		if err != nil {
			return N.NotFound()
		}
		if len(galleries) != 1 {
			return N.NotFound()
		}
		data, err := json.Marshal(galleries[0])
		if err != nil {
			return N.NotFound()
		}
		N.Write(data)
		return 200
	} else {
		return N.NotFound()
	}
}

func (G Gallery) Post(N NetReq) int {
	if !N.IsAdmin() {
		return N.Error("Unauthorized", 403)
	}

	N.R.ParseForm()
	gn := N.R.Form["galleryname"]
	if len(gn) != 1 {
		return N.Error("Missing galleryname field", 400)
	}

	galleryname := gn[0]
	dataStore := util.GetMetadataValue(N.DB, "dataStore")
	imageStore := util.GetMetadataValue(N.DB, "imageStore")

	if imageStore == "" || dataStore == "" {
		return N.Error("No Space for data or images", 500)
	}

	localpath := util.MakeFriendlyPath(galleryname)
	for web.FExists(imageStore + "/" + localpath) {
		localpath = localpath + "0"
	}

	err := os.Mkdir(imageStore + "/" + localpath, os.ModePerm)
	if err != nil {
		return N.Error("Failed to make image store: "+localpath, 500)
	}
	err = os.Mkdir(dataStore + "/" + localpath, os.ModePerm)
	if err != nil {
		return N.Error("Failed to make data store: "+localpath, 500)
	}

	err = os.Symlink("../../static/placeholder.png", imageStore + "/" + localpath + "/480placeholder.png")
	if err != nil {
		return N.NotFound()
	}

	var gallery generated.Gallery
	gallery.Name = galleryname
	gallery.Path = localpath
	gallery.Splash = "placeholder.png"
	gallery.Lat = 0
	gallery.Lon = 0
	generated.InsertGalleryTable(N.DB, gallery)
	
	return N.OK()
}

func (G Gallery) Put(N NetReq) int {
	if !N.IsAdmin() {
		return N.Error("Unauthorized", 403)
	}

	if len(N.Url) != 1 {
		return N.NotFound()
	}

	N.R.ParseForm()
	props := []string{"name", "splash", "lat", "lon", "hasgpx"}
	for _, prop := range props {
		gn := N.R.Form[prop]
		if len(gn) == 1 {
			err := generated.UpdateGalleryTable(N.DB, prop, gn[0], map[string]interface{}{
				"Path": N.Url[0],
			})
			if err != nil {
				return N.Error("Failed to change property " + prop, 500)
			}
		}
	}
	return N.OK()
}

func (G Gallery) Delete(N NetReq) int {
	return N.NotFound()
}

func (G Gallery) Patch(N NetReq) int {
	return N.NotFound()
}

func (G Gallery) Head(N NetReq) int {
	return N.NotFound()
}

func (G Gallery) ResourceName() string {
	return "gallery"
}
