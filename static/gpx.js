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
		for (var j=0; j<step+1; j++) {
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
			strokeWeight: getSpeedWeight(speed, min, max)
		}))
	}
	cb_line_center(lines, bounds.getCenter())
}

function getSpeed(trkpt) {
    if (!trkpt) {
        return 0;
    }
	return Number(trkpt.getElementsByTagName('speed')[0].childNodes[0].nodeValue);
}

function hslToHex(h, s, l) {
	h /= 360;
	s /= 100;
	l /= 100;
	let r, g, b;
	if (s === 0) {
		r = g = b = l; // achromatic
	} else {
		const hue2rgb = (p, q, t) => {
			if (t < 0) t += 1;
			if (t > 1) t -= 1;
			if (t < 1 / 6) return p + (q - p) * 6 * t;
			if (t < 1 / 2) return q;
			if (t < 2 / 3) return p + (q - p) * (2 / 3 - t) * 6;
			return p;
		};
		const q = l < 0.5 ? l * (1 + s) : l + s - l * s;
		const p = 2 * l - q;
		r = hue2rgb(p, q, h + 1 / 3);
		g = hue2rgb(p, q, h);
		b = hue2rgb(p, q, h - 1 / 3);
	}
	const toHex = x => {
		const hex = Math.round(x * 255).toString(16);
		return hex.length === 1 ? '0' + hex : hex;
	};
	return `#${toHex(r)}${toHex(g)}${toHex(b)}`;
}

function getSpeedWeight(speed, min, max) {
    s = 8 * (speed - min)
    s /= (max - min)
    return 1+s
}

function getSpeedColor(speed, min, max) {
	red = 255 * (speed-min)
	red /= (max-min)
	return hslToHex(red, 100, 50);
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










function trackElementToLatLng(elem) {
    return {
        lat: Number(elem.getAttribute('lat')),
        lng: Number(elem.getAttribute('lon'))
    }
}

function getColoredTrack2(track) {
    var lines = [];
    track = track.children;

    var max = getSpeed(track[0]);
    var min = max;

    var step = 10;

    for (var i=0; i<track.length; i+=step) {
        speed = 0;
        for (var j=0; j<step+1; j++) {
            speed += getSpeed(track[i+j]);
        }
        if (speed < min) {
            min = speed;
        }
        if (speed > max) {
            max = speed;
        }
    }
    for (var i=0; i<track.length; i+=step) {
        speed = 0;
        path = [];
        for (var j=0; j<step+1; j++) {
            speed += getSpeed(track[i+j]);
            if (track[i+j] && (j % 2 == 0)) {
                path.push(trackElementToLatLng(track[i+j]))
            }
        }
        lines.push(new google.maps.Polyline({
            path: path,
            geodesic: false,
            strokeColor: getSpeedColor(speed, min, max),
            strokeWeight: getSpeedWeight(speed, min, max),
            strokeOpacity: 1.0,
        }))
    }
    return lines;
}

function gpxRaw2Path(gpxdata) {
	var trksegs = gpxdata.getElementsByTagName("trkseg");
	track = []
	for (var i = 0; i < trksegs.length; i++) {
		track = track.concat(getColoredTrack2(trksegs[i]));
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

function path2LatLon(path) {
	var bounds = new google.maps.LatLngBounds();

	for (var i = 0; i < path.length; i++) {
		bounds.extend(path[i].getPath().getAt(0))
	}

	return bounds
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
