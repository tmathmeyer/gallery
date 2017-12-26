bin:
	cd database && python generate.py
	go build -o server main.go

run:
	cd database && python generate.py
	go run main.go

init:
	rm -f live.sqlite
	cp -n setup.template.go setup.go
	vim setup.go
	go run setup.go