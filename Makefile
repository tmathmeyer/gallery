

all: backend_ frontend_
	mkdir -p rundir
	mkdir -p rundir/static
	mkdir -p rundir/templates
	cp exif/bin/photosphere rundir/
	cp backend/server rundir/
	cp frontend/colors.yaml rundir/
	cp frontend/static/* rundir/static/
	cp -r frontend/templates/* rundir/templates/

setup: all
	make -C backend/ init
	mv backend/live.sqlite rundir/


#build the frontend code
frontend_:
	make -C frontend

# Build the backend binary
backend_: exif_
	make -C backend/

# build the exif project
exif_:
	git submodule init
	git submodule update
	make -C exif/

clean:
	make -C frontend clean
	make -C backend clean
	rm -rf rundir