all:
	make -C typescript
	mkdir static
	cp -r css/* static/
	cp -r icons/* static/
	cp -r js/* static/
	cp -r libraries/* static/
	cp -r templates/* static/
	cp -r typescript/*.js static/

clean:
	make -C typescript clean
	rm -rf static