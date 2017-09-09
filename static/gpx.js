function getTrackFromUrl(url, cb_line_center) {
	$.ajax({url: url, dataType: "xml",
		success: function(data) {
			var trksegs = data.getElementsByTagName("trkseg");
			for (var i = 0; i < trksegs.length; i++) {
				getTrackAndMidpoint(trksegs[i], cb_line_center);
			}
		}, failure: function(err){}
	})
}

function getTrackAndMidpoint(track, cb_line_center) {
	var bounds = new google.maps.LatLngBounds();
	var lines = [];

	var step = 50;
	var min = 9999999;
	var max = 0;
	for (var i=0; i<track.children.length; i+=step) {
		var speed = 0;
		for (var j=0; j<5; j++) {
			A = track.children[i+j]
			if (A) {
				speed += getSpeed(A);
			}
		}
		if (speed < min) {
			min = speed;
		}
		if (speed > max) {
			max = speed;
		}
	}
	for (var i = 0; i < track.children.length; i+=step) { 
		var path = [];
		var speed = 0;
		for (var j=0; j<step+1; j++) {
			A = track.children[i+j]
			if (A) {
				speed += getSpeed(A);
				if (j%10==0) {
					path.push({lat: Number(A.getAttribute('lat')), lng: Number(A.getAttribute('lon'))})
				}
			}
		}
		bounds.extend(new google.maps.LatLng(path[0]))
		lines.push(new google.maps.Polyline({
			path: path,
			geodesic: false,
			strokeColor: getSpeedColor(speed, min, max),
			strokeOpacity: 1.0,
			strokeWeight: 2
		}))
	}
	cb_line_center(lines, bounds.getCenter())
}

function getSpeed(trkpt) {
	return Number(trkpt.getElementsByTagName('speed')[0].childNodes[0].nodeValue);
}

function getSpeedColor(speed, min, max) {
	color = 255 * (speed-min)
	color /= (max-min);
	str =  '#' + ("0"+(Number(color).toString(16))).slice(-2).toUpperCase() + "0000"
	return str
}

function getGPXInfo(url, name, map, bounds) {
	getTrackFromUrl(url, function(lines, point) {
		for (var i=0; i<lines.length; i++) {
			lines[i].setMap(map);
		}
		bounds.extend(point);
		marker = new google.maps.Marker({
			position: point,
			map: map,
			title: name
		});
	});
}














function getColoredTrack(track) {
	var lines = [];

	var step = 50;
	var min = 9999999;
	var max = 0;
	for (var i=0; i<track.children.length; i+=step) {
		var speed = 0;
		for (var j=0; j<5; j++) {
			A = track.children[i+j]
			if (A) {
				speed += getSpeed(A);
			}
		}
		if (speed < min) {
			min = speed;
		}
		if (speed > max) {
			max = speed;
		}
	}
	for (var i = 0; i < track.children.length; i+=step) { 
		var path = [];
		var speed = 0;
		for (var j=0; j<step+1; j++) {
			A = track.children[i+j]
			if (A) {
				speed += getSpeed(A);
				if (j%10==0) {
					path.push({lat: Number(A.getAttribute('lat')), lng: Number(A.getAttribute('lon'))})
				}
			}
		}
		lines.push(new google.maps.Polyline({
			path: path,
			geodesic: false,
			strokeColor: getSpeedColor(speed, min, max),
			strokeOpacity: 1.0,
			strokeWeight: 2
		}))
	}
	return lines;
}

function gpxRaw2Path(gpxdata) {
	var trksegs = gpxdata.getElementsByTagName("trkseg");
	track = []
	for (var i = 0; i < trksegs.length; i++) {
		track = track.concat(getColoredTrack(trksegs[i]));
	}
	return track
}

function readGpxPath(path, cb) {
	$.ajax({url: '/d/'+path+'/gpx', dataType: "xml",
		success: function(data) {
			cb(gpxRaw2Path(data))
		}
	}).fail(function(err) {
		cb(null);
		return false;
	})
}

function path2LatLon(path, result) {
	var bounds = new google.maps.LatLngBounds();

	for (var i = 0; i < path.length; i++) {
		bounds.extend(path[i].getPath().getAt(0))
	}

	center = bounds.getCenter()
	result.lat = center.lat()
	result.lon = center.lng()
}

function getLocation(path, lat, lon, cb) {
	lat = parseFloat(lat)
	lon = parseFloat(lon)
	res = {'gpx': null, 'lat': null, 'lon': null}
	readGpxPath(path, function(path) {
		if (path) {
			res['gpx'] = path
		}
		if (path && lat==0 && lon==0) {
			path2LatLon(path, res)
		}
		if (lat != 0 || lon != 0) {
			res['lat'] = lat
			res['lon'] = lon
		}
		cb(res)
	})
}

function displayData(info, map) {
	console.log(info)
	for (var p in info) {
		e = info[p]
		if (e.gpx != null) {
			renderGpx(e.gpx, map)
		}
		if (e.lat != null) {
			drawPoint(e.lat, e.lon, map);
		}
	}
}

function renderGpx(gpx, map) {
	for(var i=0; i<gpx.length; i++) {
		gpx[i].setMap(map);
	}
}

function drawPoint(lat, lon, map) {
	new google.maps.Marker({
		position: {lat: lat, lng: lon}
	}).setMap(map);
}