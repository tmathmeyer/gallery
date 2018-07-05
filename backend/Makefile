
bin: database_
	go build -o server main.go

database_:
	cd database && python generate.py

init:
	cp -n setup.template.go setup.go
	$$EDITOR setup.go
	go run setup.go

clean:
	rm -rf database/generated
	rm -rf server
	rm -rf *.sqlite