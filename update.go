package main

import (
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"./database/util"
	"./database/generated"
	"database/sql"
    "path/filepath"
    "strings"
)

func queryOldMetadata(db *sql.DB, key string) string {
	rows, err := db.Query("select value from metadata where key is '" + key + "'")
	if err != nil {
		return ""
	}
	defer rows.Close()
	for rows.Next() {
		var Value string
		err = rows.Scan(&Value)
		if err != nil {
			return ""
		}
		return Value
	}
	return ""
}


func main() {
	db, err := generated.OpenDatabase("live.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	db2, err := sql.Open("sqlite3", "old.sqlite")
	defer db.Close()

	newDataLocation := util.GetMetadataValue(db, "dataStore")
	newPhotosLocation := util.GetMetadataValue(db, "imageStore")
	log.Println(newDataLocation)
	log.Println(newPhotosLocation)


	oldPhotosLocation := queryOldMetadata(db2, "galleryData")
	log.Println(oldPhotosLocation)




	photos := make(map[string][]string)
	valid_gal := make(map[string]int)
	err = filepath.Walk(oldPhotosLocation, func(path string, f os.FileInfo, err error) error {
		parts := strings.SplitN(path, "/", 3)
		if (len(parts) == 3) {
			gallery, img := parts[1], parts[2]


			photos[gallery] = append(photos[gallery], img)
			if strings.Contains(img, "/") {
				valid_gal[gallery] = 2
			}


			log.Println(parts)
		}
		return nil
	})








}
