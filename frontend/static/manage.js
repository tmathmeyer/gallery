function getCookie(name) {
  var value = "; " + document.cookie;
  var parts = value.split("; " + name + "=");
  if (parts.length == 2) return parts.pop().split(";").shift();
}

function elem_prop(elem, prop) {
	return elem.getElementsByClassName("__prop_"+prop)[0].value;
}

function sendData(type, gallery, special) {
	url = '/api/v_dev/gallery/'
	if (type == 'PUT') {
		url += gallery
	}
	if (typeof special !== 'undefined') {
		url = special
	}
	return function(data, success, xhr) {
		req = {
			url: url,
			type: type,
			data: data,
			headers: {"Authorization": "Bearer "+getCookie("jwt")}
		}
		if (typeof success !== 'undefined') {
			req['success'] = success
		}
		if (typeof xhr !== 'undefined') {
			req['xhr'] = xhr
		}
		$.ajax(req)
	}
}

function createImageHolder(file, cb) {
	photo = document.createElement('div')
	img = document.createElement('img')
	progress = document.createElement('progress')
	thb = document.createElement('div')

	progress.classList.add('hidden')
	photo.classList.add('uploading-photo')
	thb.classList.add('thumbnailsetter')
	thb.classList.add('hidden')
	thb.innerHTML = 'Set Thumbnail'
	thb.onclick = setThumbnail

	img.src = '/static/loading.gif'
	photo.appendChild(img)
	photo.appendChild(progress)
	photo.appendChild(thb)

	photo.setAttribute('data-gallery', $("#gallerynamesubmit").attr("path"))

	cb({
		image: file,
		element: photo
	})
}

function setThumbnail() {
	thmb = document.getElementsByClassName('thumbnailsetter')
	for (i=0; i<thmb.length; i++) {
		thmb[i].classList.add('hidden')
	}
	parent = this.parentElement
	gallery = parent.getAttribute('data-gallery')
	name = parent.getAttribute('data-name')

	sendData('PUT', gallery)({splash:name}, function() {
		location.reload()
	})
}

function change_password(id) {
	pass = prompt("Please Enter a New Password (1/2):", "")
	if (pass == null) {
		return
	}
	if (pass.length < 5) {
		alert("Password must be >= 5 characters")
		return
	}

	repeat = prompt("Please Repeat New Password (2/2):", "")
	if (pass != repeat) {
		alert("Passwords do not match, Please Try Again")
		change_password()
		return
	}

	data = {"password": pass}
	if (id != null) {
		data['id'] = id
	}

	sendData('PUT', 0, '/api/v_dev/user/')(data, function(status) {
		alert('Password Changed!')
	})
}

function add_new_user() {
	username = prompt("Please Enter a Username (1/3):", "")
	if (username == null) {
		return
	}
	pass = prompt("Please Ender a Password for "+username+" (2/3):", "")
	if (pass == null) {
		return
	}
	if (pass.length < 5) {
		alert("Password must be >= 5 characters")
		return
	}
	repeat = prompt("Please Repeat New Password (2/2):", "")
	if (pass != repeat) {
		alert("Passwords do not match, Please Start Over.")
		add_new_user()
		return
	}
	sendData('POST', 0, '/api/v_dev/user/')({
		'username': username,
		'password': pass
	}, function(status) {
		show_user_management(()=>show_user_management())
	})
}

function span(ct, clazz){
	s = document.createElement('span')
	s.className = clazz
	if (typeof ct == "string") {
		s.innerHTML = ct;
	} else {
		s.appendChild(ct)
	}
	return s
}

function createTable(table, data, columnGenerators, cb) {
	header = table.createTHead().insertRow();
	for (n in columnGenerators) {
		header.insertCell().appendChild(span(columnGenerators[n][0](), 'header'))
	}

	body = table.createTBody()
	ct = 0;
	data.forEach(function(e) {
		row = body.insertRow()
		for (n in columnGenerators) {
			row.insertCell().appendChild(span(columnGenerators[n][1](e), 'user-row '+n))
		}
		ct ++;
		if (ct == data.length) {
			if (cb){(cb())}
		}
	})
}

function delete_user(username) {
	sendData('DELETE', 0, '/api/v_dev/user/'+username)({},function(status) {
		show_user_management(()=>show_user_management())
	})
}

function makeTableLink(text, call, args) {
	sp = span(text, 'pseudolink')
	sp.onclick = function() {
		call.apply(null, args)
	}
	return sp
}

function populate_user_management(cb) {
	sendData('GET', 0, '/api/v_dev/user/')({}, function(data) {
		table = document.getElementById("usertable")
		createTable(table, data, {
			'Name': [()=>'Name', (n)=>n.Name],
			'Passhash': [()=>'Change Password', (n)=>makeTableLink('change', change_password, [n.Id])],
			'Admin': [()=>'Delete User', (n)=>n.Admin==1?'':makeTableLink('delete', delete_user, [n.Name])]
		}, cb)
	})
}

function show_user_management(cb) {
	$('#user_management').toggleClass('hidden')
	if (document.getElementById("usertable").innerHTML.trim() == "") {
		populate_user_management(cb);
	} else {
		document.getElementById("usertable").innerHTML = ""
	}
}

function populateImagesWithData(img_el, cb) {
	img = img_el['image']
	ele = img_el['element']
	ele.querySelectorAll('img')[0].src = '/img/'+img.Gallery+'/'+img.Name+'/TN'
	ele.setAttribute('data-name', img.Name)
	cb(img_el)
}

function makePhotoSquares(photos) {
	callSynchronously([
		createImageHolder,
		renderTemporaryImage,
		populateImagesWithData
	], photos)
}

function queueUpload(files) {
	callSynchronously([
		createImageHolder,
		renderTemporaryImage,
		populateImageHolder,
		uploadFileForImage
	], files)
}

function load_gallery_for_changes(elem) {
	if (!elem) {
		return;
	}
	var name = elem_prop(elem, "name");
	var path = elem_prop(elem, "path");
	var splash = elem_prop(elem, "splash");
	document.location.hash = elem.id
	$("#gallerynamefield").val(name)
	$("#gallerynamesubmit").attr("path", path)
	$("#galleryeditor").removeClass("hidden")

	document.getElementById("photos").innerHTML = ''
	sendData('GET', null, '/api/v_dev/images/'+path)(null, makePhotoSquares)
}

function syncMap(iterable, asyncMapper, acceptor) {
	(function(map) {
		map(0, map, [])
	})(function(index, doNext, result) {
		if (iterable) {
			asyncMapper(iterable[index], function(B) {
				result.push(B)
				if (index+1 >= iterable.length) {
					acceptor(result)
				} else {
					doNext(index+1, doNext, result)
				}
			})
		} else {
			acceptor(result)
		}
	})
}

function callSynchronously(funcs, data, cb) {
	f = funcs.shift()
	if (f) {
		syncMap(data, f, function(d) {
			callSynchronously(funcs, d, cb)
		})
	} else if (cb) {
		cb(data)
	}
}

function populateImageHolder(img_el, element_acceptor) {
	var reader = new FileReader()
	reader.onload = function (e) {
		img = img_el['element'].querySelectorAll('img')[0]
		img.src = e.target.result
		element_acceptor(img_el)
	};
	reader.readAsDataURL(img_el['image'])
}

function renderTemporaryImage(element, ea) {
	document.getElementById("photos").appendChild(element['element'])
	ea(element)
}

function uploadFileForImage(img_el, ea) {
	form = new FormData()
	form.append('newimage', img_el['image'])
	image = img_el['element']
	image.querySelectorAll('progress')[0].classList.remove('hidden')
	$.ajax({
		url: '/api/v_dev/images/'+$("#gallerynamesubmit").attr("path"),
		type: 'POST',
		data: form,
		cache: false,
		contentType: false,
		processData: false,
		success: function(data) {
			image.setAttribute('data-name', data.trim())
			image.querySelectorAll('progress')[0].classList.add('hidden')
			ea(img_el)
		},
		xhr: function() {
			var myXhr = $.ajaxSettings.xhr();
			if (myXhr.upload) {
				myXhr.upload.addEventListener('progress', function(e) {
					if (e.lengthComputable) {
						prog = image.querySelectorAll('progress')[0]
						prog.setAttribute('value', e.loaded)
						prog.setAttribute('max', e.total)
					}
				} , false);
			}
			return myXhr;
		}
	})
}

function uploadGPX(map) {
	form = new FormData()
	form.append('gpx', document.getElementById('gpxupload_hidden').files[0])
	$.ajax({
		url: '/api/v_dev/gallery/'+$("#gallerynamesubmit").attr("path") + '/gpx',
		type: 'POST',
		data: form,
		cache: false,
		contentType: false,
		processData: false,
		success: function(data) {
			render_gpx(map, {'lat': 0, 'lon': 0})
		}
	})
}

function get_gallery_name() {
	return document.location.hash.slice(1);
}

$(document).ready(function() {

	if (document.location.hash) {
		load_gallery_for_changes(document.getElementById(get_gallery_name()))
	}

	$("#createButton").click(function(e) {
		newname = prompt("Gallery name:", "")
		if (newname) {
			sendData('POST')({galleryname: newname}, function() {
				location.reload()
			})
		}
	});

	$("#thumbnail_toggle").click(function(e) {
		thmb = document.getElementsByClassName('thumbnailsetter')
		for (i=0; i<thmb.length; i++) {
			thmb[i].classList.toggle('hidden')
		}
	});

	$('#imageupload').click(function(){
		$('#imageupload_hidden').trigger('click');
	});

	$('#change_password').click(function(){
		change_password();
	})

	$('#manage_users').click(function() {
		show_user_management();
	})

	$('#add_new_user').click(function() {
		add_new_user();
	})

	$('#imageupload_hidden').change(function(e) {
		queueUpload(e.target.files)
	})

	$(".albummanager").click(function(){
		load_gallery_for_changes(this)
	});

	$('#upload-gpx').click(function(){
		$('#gpxupload_hidden').trigger('click');
	});

	$("#gallerynamesubmit").click(function() {
		name = $("#gallerynamefield").val()
		path = $("#gallerynamesubmit").attr("path")
		if (name) {
			sendData('PUT', path)({name: name}, function() {
				location.reload()
			})
		}
	});
});

function render_gpx(map, location_data) {
	$.ajax({url: '/api/v_dev/gallery/'+get_gallery_name()+'/gpx', dataType: "xml",
		success: function(data) {
			mapGPX = gpxRaw2Path(data)
			renderGpx(mapGPX, map)
			google.maps.event.trigger(map, 'resize')
			map.fitBounds(path2LatLon(mapGPX))
		},
		error: function(d){
			render_marker(map, location_data.lat, location_data.lon)
		}
	});
}

function render_marker(map, lat, lng) {
	center = new google.maps.LatLng(lat, lng)
	map.setZoom(7)
	map.marker = new google.maps.Marker({
		position: center,
		map: map
	})
	google.maps.event.trigger(map, 'resize')
	map.setCenter(center)
}

function render_map_panel(map, getter) {
	getter({}, function(location_data) {
		console.log(location_data)
		if (location_data.hasgpx) {
			render_gpx(map, location_data)
		} else {
			render_marker(map, location_data.lat, location_data.lon)
		}
	})
}

function init_map() {
	(function(element){
		map = new google.maps.Map(element, {
			center: {lat: 0, lng: 0},
			zoom: 7,
			mapTypeId: 'terrain',
			disableDefaultUI: true
		});
		map.marker = null;
		map.is_setting_location = false

		map_element = $(element)
		map_configuration = $("#map_configuration")

		function add_one_time_click_listener() {
			$('#set-location').clickonce(function(e) {
				map.is_setting_location = true
				$('#set-location').toggleClass('highlight-toggle')
			})
		}

		map.addListener('click', function(c){
			if (map.is_setting_location) {
				map.is_setting_location = false
				sendData('PUT', get_gallery_name())({
					'lat': c.latLng.lat(),
					'lon': c.latLng.lng()
				}, function() {
					if (map.marker) {
						map.marker.setMap(null)
					}
					map.marker = new google.maps.Marker({
						position: c.latLng,
						map: map
					})
					map.panTo(c.latLng)
					$('#set-location').toggleClass('highlight-toggle')
				})
				add_one_time_click_listener()
			}
		})

		$('.locationconfig-display').click(function() {
			map_configuration.toggleClass('hidden')
			map_configuration.append(map_element)
			render_map_panel(map, sendData('GET', 0, '/api/v_dev/gallery/'+get_gallery_name()+'/location'))
		})

		$('.locationconfig-destroy').click(function() {
			map_configuration.toggleClass('hidden')
			$('#blackhole').append(map_element)
		})

		$('#gpxupload_hidden').unbind()
		$('#gpxupload_hidden').change(function() {
			uploadGPX(map);
		})

		add_one_time_click_listener()

	})(document.getElementById("map-conf-map"))
}

