function getCookie(name) {
  var value = "; " + document.cookie;
  var parts = value.split("; " + name + "=");
  if (parts.length == 2) return parts.pop().split(";").shift();
}

function elem_prop(elem, prop) {
	return elem.getElementsByClassName("__prop_"+prop)[0].value;
}

function sendData(type, gallery, special) {
	url = '/api/gallery'
	if (type == 'PUT') {
		url += '/'+gallery
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
	], photos, function(){})
}

function queueUpload(files) {
	callSynchronously([
		createImageHolder,
		renderTemporaryImage,
		populateImageHolder,
		uploadFileForImage
	], files, console.log)
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
	sendData('GET', null, '/api/image/'+path)(null, makePhotoSquares)
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
	} else {
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
		url: '/api/upload/'+$("#gallerynamesubmit").attr("path"),
		type: 'POST',
		data: form,
		cache: false,
		contentType: false,
		processData: false,
		success: function(data) {
			console.log(data)
			image.setAttribute('data-name', data)
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

$(document).ready(function() {

	if (document.location.hash) {
		load_gallery_for_changes(document.getElementById(document.location.hash.slice(1)))
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

	$('#imageupload_hidden').change(function(e) {
		queueUpload(e.target.files)
	})

	$(".albummanager").click(function(){
		load_gallery_for_changes(this)
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
