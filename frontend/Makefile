all:
	make -C typescript
	mkdir -p static
	cp -r css/* static/
	cp -r icons/* static/
	cp -r js/* static/
	cp -r libraries/* static/
	mv typescript/*.js static/

clean:
	make -C typescript clean
	rm -rf static