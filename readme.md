Gallery
=======

A really simple host-it-yourself photo gallery project

## Dependancies
 - python
 - typescript
 - libvips
 - imagemagick
 - golang
   - `go get github.com/dgrijalva/jwt-go`
   - `go get github.com/mattn/go-sqlite3`
   - `go get golang.org/x/crypto/bcrypt`
   - `go get gopkg.in/yaml.v2`


## Installation
 - See dependancies section for what to install
 - run `make setup` to edit a preferences file and get things set up.
   - Note: The private key field is _not_ a password. You never need to remember it, and you should generate a random base64 string, ideally more than 40 characters in length.
 - run `cd runfiles && ./server`
 - navigate to `localhost:7923` and press the fingerprint icon in the top right to start adding photos.
