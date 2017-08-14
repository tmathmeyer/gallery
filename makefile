all:
	go build -o server gallery_public.go database.go gallery_admin.go authentication.go

setup: clean
	go run setup.go database.go

clean:
	rm -rf server
	rm -f live.sqlite