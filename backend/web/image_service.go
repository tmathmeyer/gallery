package web

import (
	"database/sql"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"../database/util"
	"log"
)

// GET /img/:gallery/:photo/:size
func ImageRawHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url_data := GetPathParts(r, "/img/", 3)
		if len(url_data) != 3 {
			http.NotFound(w, r)
			return
		}

		gallery := url_data[0]
		photo := url_data[1]
		size := url_data[2]

		if strings.HasPrefix(gallery, ".") || strings.HasPrefix(photo, ".") {
			http.NotFound(w, r)
			return
		}

		res := ""
		switch size {
		case "O":
			res = ""
		case "UHD":
			res = "3840"
		case "WQHD":
			res = "2560"
		case "HD1":
			res = "1920"
		case "HD7":
			res = "1280"
		case "VGA":
			res = "640"
		case "TN":
			res = "480"
		case "QVGA":
			res = "320"
		}

		galleryFsLocation := util.GetMetadataValue(db, "imageStore")

		// ex: photos/MyTrip/640img0001.jpg
		imagePath := fmt.Sprintf("%s/%s/%s%s", galleryFsLocation, gallery, res, photo)

		if FExists(imagePath) {
			http.ServeFile(w, r, imagePath)
			return
		}

		if url_data[0] != "O" && res != "" {
			original_path := fmt.Sprintf("%s/%s/%s", galleryFsLocation, gallery, photo)
			pattern := fmt.Sprintf("%s%s%s", res, "%s", filepath.Ext(original_path))
			resizeImage(original_path, pattern, res)
			if FExists(imagePath) {
				http.ServeFile(w, r, imagePath)
				return
			}
		}

		log.Println(imagePath)

		http.NotFound(w, r)
	})
}

func resizeImage(original_path string, pattern string, size string) {
	cmd := exec.Command("vipsthumbnail", "-s", size, "-o", pattern, original_path)
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		fmt.Printf("vipsthumbnail -s %s -o %s %s\n", size, pattern, original_path)
	}
}
