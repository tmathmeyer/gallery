Gallery
=======

A really simple host-it-yourself photo gallery project

## Dependancies
 - `go get github.com/dgrijalva/jwt-go`
 - `go get github.com/mattn/go-sqlite3`
 - `go get golang.org/x/crypto/bcrypt`

## Installation
 - See dependancies section for what to install
 - Copy `setup.go.template` to `setup.go`
 - Edit `setup.go` with your desired information
   - Note: The private key field is _not_ a password. You never need to remember it, and you should generate a random base64 string, ideally more than 40 characters in length.
 - run `make setup`
 - run `make`
 - run the `server` binary
 - navigate to `localhost:8081/` and press the fingerprint icon in the top right to start adding photos.