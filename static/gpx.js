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
	console.log('finished loop 1')
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
		console.log(lines)
		console.log(point)
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








/*
function parseGPX(data, map) {
	console.log(data)
	var trksegs = data.getElementsByTagName("trkseg");
	for (var i = 0; i < trksegs.length; i++) {   
		drawtrack(trksegs[i], map);
	}
}


function drawtrack(track, map) {
	var bounds = new google.maps.LatLngBounds();
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
		line = new google.maps.Polyline({
			path: path,
			geodesic: false,
			strokeColor: getSpeedColor(speed, min, max),
			strokeOpacity: 1.0,
			strokeWeight: 2
		})
		line.setMap(map)
	}
	map.fitBounds(bounds)

	google.charts.load('current', {'packages':['corechart']});
	google.charts.setOnLoadCallback(drawChart);

	function drawChart() {
		var table = [
			['Time', 'Elevation',]
		]
		for (var i=0; i<track.children.length; i++) {
			elev = Number(track.children[i].getElementsByTagName('ele')[0].childNodes[0].nodeValue);
			date = new Date(track.children[i].getElementsByTagName('time')[0].childNodes[0].nodeValue);
			time = twelveHour(date.getHours())+":"+addZero(date.getMinutes())+AmPm(date.getHours())
			table.push([time, elev])
		}
		var data = google.visualization.arrayToDataTable(table);

		var options = {
			title: 'Elevation over time',
			curveType: 'function',
			legend: { position: 'bottom' }
		};

		var chart = new google.visualization.LineChart(document.getElementById('elevgraph'));
		chart.draw(data, options);
	}
}

function twelveHour(time) {
	if (time > 12) {
		return time-12;
	}
	return time;
}

function AmPm(time) {
	if (time > 11) {
		return 'PM'
	}
	return 'AM'
}

function addZero(mins) {
	if (mins < 10) {
		return '0'+mins
	}
	return mins
}
*/