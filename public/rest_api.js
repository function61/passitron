
function rest_search(search) {
	return $.getJSON('/secrets?search=' + encodeURIComponent(search));
}

function rest_sshkeys(search) {
	return $.getJSON('/secrets?sshkey=y');
}

function rest_byFolder(id) {
	return $.getJSON('/folder/' + id);
}

function rest_exposedCred(id) {
	return $.getJSON('/secrets/' + id + '/expose');
}

function rest_credentialById(id) {
	return $.getJSON('/secrets/' + id);
}

function restDefaultErrorHandler(xhr) {
	if (xhr.responseJSON && xhr.responseJSON.error_code) {
		if (xhr.responseJSON.error_code === 'database_is_sealed') {
			if (confirm('Error: you need to unseal the database first. Do that?')) {
				invokeCommand('UnsealRequest');
			}
		} else {
			alert(xhr.responseJSON.error_code + ': ' + xhr.responseJSON.error_description);
		}

		return;
	}

	alert('Unknown REST error, logged in console');

	console.error(xhr);
}

