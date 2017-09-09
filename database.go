package main

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"log"
	"os"
)

type Gallery struct {
	Name   string
	Path   string
	Lat    float64
	Lon    float64
	Splash string
}

type Photo struct {
	Name    string
	Descr   string
	Gallery string
}

func getPhotosByGallery(db *sql.DB, galleryID string) ([]Photo, error) {
	rows, err := db.Query("select * from photos where gallery is \"" + galleryID + "\"")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var Photos []Photo
	for rows.Next() {
		var description string
		var name string
		var gallery string
		err = rows.Scan(&name, &description, &gallery)
		if err != nil {
			return nil, err
		}
		Photos = append(Photos, Photo{
			Name:    name,
			Descr:   description,
			Gallery: gallery,
		})
	}
	return Photos, nil
}

func getGalleries(db *sql.DB) ([]Gallery, error) {
	rows, err := db.Query("select path, name, lat, lon, splash from galleries")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var Galleries []Gallery
	for rows.Next() {
		var name string
		var path string
		var lat float64
		var lon float64
		var splash string
		err = rows.Scan(&path, &name, &lat, &lon, &splash)
		if err != nil {
			return nil, err
		}
		Galleries = append(Galleries, Gallery{
			Name:   name,
			Path:   path,
			Lat:    lat,
			Lon:    lon,
			Splash: splash,
		})
	}
	return Galleries, nil
}

func getPhotosByGalleryAndName(db *sql.DB, galleryID string, photoName string, photo *Photo) error {
	rows, err := db.Query("select * from photos where gallery is \"" + galleryID + "\" and name is \"" + photoName + "\"")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var description string
		var name string
		var gallery string
		err = rows.Scan(&name, &description, &gallery)
		if err != nil {
			return err
		}
		photo.Name = name
		photo.Descr = description
		photo.Gallery = gallery
		return nil
	}
	return errors.New("Cannot get photo")
}

func getMetadataValue(db *sql.DB, metakey string) string {
	rows, err := db.Query("select value from metadata where key is \"" + metakey + "\"")
	if err != nil {
		return ""
	}
	defer rows.Close()
	for rows.Next() {
		var data string
		err = rows.Scan(&data)
		if err != nil {
			return ""
		}
		return data
	}
	return ""
}

func getGallery(db *sql.DB, galleryID string, gallery *Gallery) error {
	rows, err := db.Query("select name, path, lat, lon, splash from galleries where path is \"" + galleryID + "\"")
	if err != nil {
		return err
	}
	defer rows.Close()
	var name string
	var path string
	var lat float64
	var lon float64
	var splash string
	for rows.Next() {
		err = rows.Scan(&name, &path, &lat, &lon, &splash)
		if err != nil {
			return err
		}
		gallery.Name = name
		gallery.Path = path
		gallery.Lat = lat
		gallery.Lon = lon
		gallery.Splash = splash
		return nil
	}
	return errors.New("Cannot get gallery")
}

func file_exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func open_database(filename string) (*sql.DB, error) {
	var fileExists = file_exists(filename)

	db, err := sql.Open("sqlite3", filename)

	if err != nil {
		return nil, err
	}

	if fileExists {
		return db, nil
	}

	create_tables := `
    CREATE TABLE galleries (
         path    TEXT PRIMARY KEY,
         name    TEXT,
         lat     REAL,
         lon     REAL,
         splash  TEXT
     );
     CREATE TABLE photos (
         name          TEXT,
         description   TEXT,
         gallery       TEXT,
         FOREIGN KEY(gallery) REFERENCES gallery(path)
     );
     CREATE TABLE admins (
     	 id			   INTEGER PRIMARY KEY AUTOINCREMENT,
     	 name		   TEXT,
     	 passhash      TEXT
     );
     CREATE TABLE metadata (
     	  key    TEXT PRIMARY KEY,
     	  value  TEXT
     );
    `

	_, err = db.Exec(create_tables)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func get_prepared_transaction(db *sql.DB, query string) (*sql.Stmt, *sql.Tx, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, nil, err
	}

	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, nil, err
	}
	return stmt, tx, nil
}

func add_photo(db *sql.DB, path string, description string, gallery string) {
	stmt, tx, err := get_prepared_transaction(db, "insert into photos(name, description, gallery) values(?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(path, description, gallery)
	if err != nil {
		log.Fatal(err)
	}

	tx.Commit()
}

func set_image_description_by_name_and_gallery(db *sql.DB, gallery string, image string, description string) error {
	query := "UPDATE photos SET description=? WHERE gallery=? AND name=?"
	stmt, tx, err := get_prepared_transaction(db, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(description, gallery, image)
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func add_gallery(db *sql.DB, name string, path string, lat float64, lon float64, splash string) {
	stmt, tx, err := get_prepared_transaction(db, "insert into galleries(name, path, lat, lon, splash) values(?, ?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, path, lat, lon, splash)
	if err != nil {
		log.Fatal(err)
	}

	tx.Commit()
}

func set_gallery_prop(db *sql.DB, path string, prop string, value string) error {
	query := "UPDATE galleries SET " + prop + "=? WHERE path=?"
	stmt, tx, err := get_prepared_transaction(db, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(value, path)
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

func change_password(db *sql.DB, name string, pass string) {
	query := "UPDATE admins SET passhash=? WHERE name=?"
	stmt, tx, err := get_prepared_transaction(db, query)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	hash, err := HashPassword(pass)
	if err != nil {
		log.Fatal(err)
	}

	_, err = stmt.Exec(name, hash)
	if err != nil {
		log.Fatal(err)
	}

	tx.Commit()
}

func add_admin(db *sql.DB, name string, pass string) {
	stmt, tx, err := get_prepared_transaction(db, "insert into admins(name, passhash) values(?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	hash, err := HashPassword(pass)
	if err != nil {
		log.Fatal(err)
	}

	_, err = stmt.Exec(name, hash)
	if err != nil {
		log.Fatal(err)
	}

	tx.Commit()
}

func add_metadata(db *sql.DB, key string, value string) {
	stmt, tx, err := get_prepared_transaction(db, "insert into metadata(key, value) values(?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(key, value)
	if err != nil {
		log.Fatal(err)
	}

	tx.Commit()
}

func is_admin(db *sql.DB, user string, pass string) bool {
	rows, err := db.Query("select passhash from admins where name is \"" + user + "\"")
	if err != nil {
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var hash string
		err := rows.Scan(&hash)
		if err != nil {
			return false
		}
		return CheckPasswordHash(pass, hash)
	}
	return false
}
