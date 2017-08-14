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
	for (var p in photos) {
		elem.appendChild(getPhotoDiv(photos[p]))
	}
}

function getPhotoDiv(photo) {
	var div = document.createElement("div");
	urlImage = "url('/i/T/"+photo.Gallery+"/"+photo.Name+"')"
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
		url: '/a/g/'+gallery,
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
			url: '/a/i/'+gallery,
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

$(document).ready(function() {
	$("#createButton").click(function(e) {
		newname = prompt("Gallery name:", "")
		if (newname) {
			$.ajax({
				url: '/a/g',
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
		var name = elem_prop(this, "name");
		var path = elem_prop(this, "path");
		var splash = elem_prop(this, "splash");

		$("#gallerynamefield").val(name)
		$("#gallerynamesubmit").attr("path", path)
		$("#galleryeditor").removeClass("hidden")


		$.ajax({
			url: '/a/i/'+path,
			type: 'GET',
			headers: {"Authorization": "Bearer " +getCookie("jwt")},
			success: function(data) {
				makePhotoSquares(data)
			}
		})


	});

	$("#gallerynamesubmit").click(function() {
		name = $("#gallerynamefield").val()
		path = $("#gallerynamesubmit").attr("path")
		if (name) {
			$.ajax({
				url: '/a/g/'+path,
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
	})

	$('#imgupload').on('click', function() {
		console.log(new FormData($('#imguploadform')[0]))
		$.ajax({
			// Your server script to process the upload
			url: '/a/u/'+$("#gallerynamesubmit").attr("path"),
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
		});
	});
});