package main

import (
    _ "github.com/mattn/go-sqlite3"
	"log"
)

func main() {
	db, err := open_database("./live.sqlite")

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	change_password(db, "{NAME}", "{PASSWORD}")
}

