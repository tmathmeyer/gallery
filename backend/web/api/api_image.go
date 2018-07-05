package api

import (
	"encoding/json"
	"io"
	"os"
	"fmt"
	"strings"
	"path/filepath"
	"os/exec"
	"../../web"
	"../../database/generated"
	"../../database/util"
)

type Image struct {}

func (I Image) Get(N NetReq) int {
	if len(N.Url) == 0 {
		return N.NotFound()
	}
	photos, err := generated.QueryPhotoTable(N.DB, map[string]interface{}{
		"Gallery": N.Url[0],
	})
	if err != nil {
		return N.Error("Failed to lookup photos", 500)
	}
	jData, err := json.Marshal(photos)
	if err != nil {
		return N.Error("Couldn't generate list of photos", 500)
	}

	N.W.Header().Set("Content-Type", "application/json")
	N.Write(jData)
	return 200
}

func testImageType(filepath string) int {
	cmd := exec.Command("./photosphere", filepath)
	err := cmd.Run()
	if err != nil {
		return 0 // Not a panoramic
	} else {
		return 1 // A panoramic / photosphere
	}
}

func (I Image) Post(N NetReq) int {
	if !N.IsAdmin() {
		return N.Error("Unauthorized", 403)
	}

	N.R.ParseMultipartForm(32 << 20)
	galleryPath := N.Url[0]
	imageStore := util.GetMetadataValue(N.DB, "imageStore")

	file, handler, err := N.R.FormFile("newimage")
	if err != nil {
		fmt.Println(err)
		return N.Error("Upload failed", 500)
	}
	defer file.Close()

	fileExtension := filepath.Ext(handler.Filename)
	filename := strings.TrimSuffix(handler.Filename, fileExtension)

	writeToPath := fmt.Sprintf("%s/%s/%s%s", imageStore, galleryPath, filename, fileExtension)

	for web.FExists(writeToPath) {
		filename = filename + "0"
		writeToPath = fmt.Sprintf("%s/%s/%s%s", imageStore, galleryPath, filename, fileExtension)
	}

	f, err := os.OpenFile(writeToPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return N.Error("Upload failed", 500)
	}
	defer f.Close()
	io.Copy(f, file)

	var photo generated.Photo
	photo.Type = testImageType(writeToPath)
	photo.Name = filename+fileExtension
	photo.Description = filename+fileExtension
	photo.Gallery = galleryPath

	generated.InsertPhotoTable(N.DB, photo)
	return N.Error(filename+fileExtension, 200)
}

func (I Image) Put(N NetReq) int {
	if !N.IsAdmin() {
		return N.Error("Unauthorized", 403)
	}

	N.R.ParseForm()
	imgs := N.R.Form["image"]
	descrs := N.R.Form["description"]

	if len(imgs) != 1 || len(descrs) != 1 {
		return N.Error("missing image or description", 400)
	}

	err := generated.UpdatePhotoTable(N.DB, "Description", descrs[0], map[string]interface{}{
		"Gallery": N.Url[0],
		"Name": imgs[0],
	});

	if err != nil {
		return N.Error("Failed to set description", 500)
	}

	return N.Error("OK", 200)
}

func (I Image) Delete(N NetReq) int {
	return N.NotFound()
}

func (I Image) Patch(N NetReq) int {
	return N.NotFound()
}

func (I Image) Head(N NetReq) int {
	return N.NotFound()
}

func (I Image) ResourceName() string {
	return "images"
}
