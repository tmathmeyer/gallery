function getCookie(name) {
  var value = "; " + document.cookie;
  var parts = value.split("; " + name + "=");
  if (parts.length == 2) return parts.pop().split(";").shift();
}

function elem_prop(elem, prop) {
	return elem.getElementsByClassName("__prop_"+prop)[0].value;
}

function makePhotoSquares(photos) {
	elem = document.getElementById("micropics")
	upload = document.getElementById("picuploader")
	elem.innerHTML = '';
	elem.appendChild(upload)
	for (var p in photos) {
		elem.appendChild(getPhotoDiv(photos[p]))
	}
}

function getPhotoDiv(photo) {
	var div = document.createElement("div");
	urlImage = "url('/img/"+photo.Gallery+"/"+photo.Name+"/TN')"
	console.log(urlImage)
	div.style.backgroundImage = urlImage;
	div.className="tilesquare"
	div.setAttribute("imageName", photo.Name);
	div.setAttribute("galleryPath", photo.Gallery);
	div.setAttribute("imageDescription", photo.Descr);

	var b1 = document.createElement("button")
	b1.onclick = function() {setThumbnail(div)}
	b1.innerHTML = "Set Thumbnail"
	
	var b2 = document.createElement("button")
	b2.onclick = function() {changeDescription(div)}
	b2.innerHTML = "Change Description"

	div.appendChild(b1)
	div.appendChild(b2)
	
	return div;
}

function setThumbnail(elem) {
	gallery = elem.getAttribute('gallerypath')
	image = elem.getAttribute('imagename')
	$.ajax({
		url: '/api/gallery/'+gallery,
		type: 'PUT',
		data: {
			splash: image
		},
		headers: {"Authorization": "Bearer " +getCookie("jwt")},
		success: function () {
			location.reload()
		}
	})
}

function changeDescription(elem) {
	gallery = elem.getAttribute('gallerypath')
	image = elem.getAttribute('imagename')
	oldDescription = elem.getAttribute('imagedescription')

	newDesc = prompt("Description", oldDescription)

	if (newDesc && newDesc != oldDescription) {
		$.ajax({
			url: '/api/image/'+gallery,
			type: 'PUT',
			data: {
				image: image,
				description: newDesc
			},
			headers: {"Authorization": "Bearer " +getCookie("jwt")},
			success: function () {
				elem.setAttribute('imageDescription', newDesc)
			}
		})
	}
}

function setMapPoint(lat, lon) {
	if (document.selection_marker) {
		document.selection_marker.setMap(null);	
	}
	document.selection_marker = new google.maps.Marker({
		position: {lat: parseFloat(lat), lng: parseFloat(lon)}
	});
	document.selection_marker.setMap(document.selection_map);
}

function load_centerpoint_location(elem) {
	var lat = elem_prop(elem, "lat");
	var lon = elem_prop(elem, "lon");
	setMapPoint(lat, lon)
}

function load_gallery_for_changes(elem) {
	if (!elem) {
		return;
	}
	var name = elem_prop(elem, "name");
	var path = elem_prop(elem, "path");
	var splash = elem_prop(elem, "splash");

	load_centerpoint_location(elem)

	document.location.hash = elem.id

	$("#gallerynamefield").val(name)
	$("#gallerynamesubmit").attr("path", path)
	$("#galleryeditor").removeClass("hidden")


	$.ajax({
		url: '/api/image/'+path,
		type: 'GET',
		headers: {"Authorization": "Bearer " +getCookie("jwt")},
		success: function(data) {
			makePhotoSquares(data)
		}
	})
}

$(document).ready(function() {

	$("#createButton").click(function(e) {
		newname = prompt("Gallery name:", "")
		if (newname) {
			$.ajax({
				url: '/api/gallery',
				type: 'POST',
				data: {
					galleryname: newname
				},
				headers: {"Authorization": "Bearer " +getCookie("jwt")},
				success: function () {
					location.reload()
				}
			})
		}
	});

	$(".albummanager").click(function(){
		load_gallery_for_changes(this)
	});

	$("#gallerynamesubmit").click(function() {
		name = $("#gallerynamefield").val()
		path = $("#gallerynamesubmit").attr("path")
		if (name) {
			$.ajax({
				url: '/api/gallery/'+path,
				type: 'PUT',
				data: {
					name: name
				},
				headers: {"Authorization": "Bearer " +getCookie("jwt")},
				success: function () {
					location.reload()
				}
			})
		}
	});

	$("#map_update").click(function() {
		path = $("#gallerynamesubmit").attr("path")
		$.ajax({
			url: '/api/gallery/'+path,
			type: 'PUT',
			data: {
				lat: document.selection_marker.position.lat(),
				lon: document.selection_marker.position.lng()
			},
			headers: {"Authorization": "Bearer " +getCookie("jwt")}
		})
	})

	$("#map_reset").click(function() {
		load_centerpoint_location(document.getElementById(document.location.hash.slice(1)))
	});

	$('#imgupload').on('click', function() {
		console.log(new FormData($('#imguploadform')[0]))
		$.ajax({
			// Your server script to process the upload
			url: '/api/upload/'+$("#gallerynamesubmit").attr("path"),
			type: 'POST',

			// Form data
			data: new FormData($('#imguploadform')[0]),

			// Tell jQuery not to process data or worry about content-type
			// You *must* include these options!
			cache: false,
			contentType: false,
			processData: false,

			// Custom XMLHttpRequest
			xhr: function() {
				var myXhr = $.ajaxSettings.xhr();
				if (myXhr.upload) {
					// For handling the progress of the upload
					myXhr.upload.addEventListener('progress', function(e) {
						if (e.lengthComputable) {
							$('progress').attr({
								value: e.loaded,
								max: e.total,
							});
						}
					} , false);
				}
				return myXhr;
			},
			success: function(data) {
				elem = document.getElementById("micropics")
				data = data.slice(0, -1)
				elem.appendChild(getPhotoDiv({
					"Gallery": $("#gallerynamesubmit").attr("path"),
					"Name":data,
					"Descr":data
				}))
				$('progress').attr({
					value: 0,
					max: 100,
				});
			}
		});
	});
});
