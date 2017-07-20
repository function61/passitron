
function rest_search(search) {
	return $.getJSON('/secrets?search=' + encodeURIComponent(search));
}

function rest_sshkeys(search) {
	return $.getJSON('/secrets?sshkey=y');
}

function rest_byFolder(id) {
	return $.getJSON('/folder/' + id);
}

function exposedCred(id) {
	return $.getJSON('/secrets/' + id + '/expose');
}

function credentialById(id) {
	return $.getJSON('/secrets/' + id);
}
