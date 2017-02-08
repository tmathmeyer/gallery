# Gallery
Just a simple gallery app written in golang.  

## Dependancies
libvips (for the vipsthumbnail binary)  
golang  
BurntSushi: toml (go get github.com/BurntSushi/toml)  
Google maps API key  

##setup
 1 copy config.toml.example to config.toml  
 2 put your gmaps api key into config.toml  
 3 make a directory called 'gallerydata'  
 4 in gallery data make a directory for each album  
 5 in each album make two directories: 'img', 'pan'  
 6 put images in to the img dir and panoramics into the pan dir  
 7 run `go build server.go`  
 8 run `server`  

##example
https://hikes.tmathmeyer.me/
