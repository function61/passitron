
function rest_search(search) {
	return $.getJSON('/accounts?search=' + encodeURIComponent(search));
}

function rest_sshkeys(search) {
	return $.getJSON('/accounts?sshkey=y');
}

function rest_byFolder(id) {
	return $.getJSON('/folder/' + id);
}

function rest_exposedCred(id) {
	return $.getJSON('/accounts/' + id + '/secrets');
}

function rest_credentialById(id) {
	return $.getJSON('/accounts/' + id);
}

function restDefaultErrorHandler(xhr) {
	if (xhr.responseJSON && xhr.responseJSON.error_code) {
		if (xhr.responseJSON.error_code === 'database_is_sealed') {
			invokeCommand('UnsealRequest');
		} else {
			alert(xhr.responseJSON.error_code + ': ' + xhr.responseJSON.error_description);
		}

		return;
	}

	alert('Unknown REST error, logged in console');

	console.error(xhr);
}

