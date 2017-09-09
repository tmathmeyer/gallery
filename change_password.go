package main

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func main() {
	db, err := open_database("./live.sqlite")

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	name := ""
	pass := ""
	change_password(db, name, pass)
	fmt.Printf("%s\n", is_admin(db, name, pass))

}
