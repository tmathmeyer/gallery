package api

import (
	"encoding/json"
	"io"
	"os"
	"fmt"
	"../../web"
	"../../database/generated"
	"../../database/util"
)

type Gallery struct {}

func (G Gallery) ListGalleries(N NetReq) int {
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
}

func (G Gallery) GetGalleryDetail(N NetReq) int {
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
}

func (G Gallery) GetGalleryDataFile(N NetReq) int {
	galleries, err := generated.QueryGalleryTable(N.DB, map[string]interface{}{
		"Path": N.Url[0],
	})
	if err != nil {
		return N.NotFound()
	}
	if len(galleries) != 1 {
		return N.NotFound()
	}

	dataFsLocation := util.GetMetadataValue(N.DB, "dataStore")
	switch N.Url[1] {
	case "gpx":
		return N.ServeFile(fmt.Sprintf("%s/%s/route.gpx", dataFsLocation, galleries[0].Path))
	case "location":
		jData, err := json.Marshal(map[string]interface{}{
			"lat": galleries[0].Lat,
			"lon": galleries[0].Lon,
			"hasgpx": galleries[0].Hasgpx,
		})
		if err != nil {
			return N.Error("failed to generate location data", 500)
		}

		N.W.Header().Set("Content-Type", "application/json")
		N.Write(jData)
		return 200
	default:
		return N.NotFound()
	}
}

func (G Gallery) Get(N NetReq) int {
	N.W.Header().Set("Content-Type", "application/json")
	switch(len(N.Url)) {
	case 0: return G.ListGalleries(N)
	case 1: return G.GetGalleryDetail(N)
	case 2: return G.GetGalleryDataFile(N)
	default: return N.NotFound()
	}
}

func (G Gallery) CreateGallery(N NetReq) int {
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

func (G Gallery) UploadGPX(N NetReq) int {
	N.R.ParseMultipartForm(32 << 15)
	galleryPath := N.Url[0]
	imageStore := util.GetMetadataValue(N.DB, "dataStore")

	file, _, err := N.R.FormFile("gpx")
	if err != nil {
		fmt.Println(err)
		return N.Error("Upload Failed", 500)
	}
	defer file.Close()

	writeToPath := fmt.Sprintf("%s/%s/route.gpx", imageStore, galleryPath)

	f, err := os.OpenFile(writeToPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return N.Error("Upload Failed", 500)
	}
	defer f.Close()
	io.Copy(f, file)

	generated.UpdateGalleryTable(N.DB, "hasgpx", 1, map[string]interface{}{
		"Path": galleryPath,
	})

	return N.OK()
}

func (G Gallery) Post(N NetReq) int {
	if !N.IsAdmin() {
		return N.Error("Unauthorized", 403)
	}
	switch(len(N.Url)) {
	case 0: return G.CreateGallery(N)
	case 2: 
	if N.Url[1] == "gpx" {
		return G.UploadGPX(N)
	}
	}
	return N.NotFound()
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
