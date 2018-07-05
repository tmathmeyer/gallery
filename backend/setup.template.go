package main

import (
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"./database/util"
	"./database/generated"
)

func main() {
	db, err := generated.OpenDatabase("live.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()


// DO NOT EDIT ABOVE HERE
	util.AddMetadata(db, "siteName", "<your domain name here>")
	util.AddMetadata(db, "gmapsApiKey", "<your google maps api key here>")
	util.AddMetadata(db, "dataStore", "<a directory to store data>")
	util.AddMetadata(db, "imageStore", "<a directory to store images>")
	util.AddMetadata(db, "secret", "<Private key>") // just some long random string!
	util.AddAdmin(db, "<admin username>", "<admin password>")
// DO NOT EDIT BELOW HERE


	os.Mkdir("../rundir/" + util.GetMetadataValue(db, "dataStore"), os.ModePerm)
	os.Mkdir("../rundir/" + util.GetMetadataValue(db, "imageStore"), os.ModePerm)
}