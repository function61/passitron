
function allCreds(search) {
	if (search) {
		return $.getJSON('/secrets?search=' + encodeURIComponent(search));
	} else {
		return $.getJSON('/secrets');
	}
}

function byFolder(id) {
	return $.getJSON('/folder/' + id);
}

/*
function credById(id) {
	return allCreds().then(function (allCredentials){
		var poop = $.Deferred();

		var matches = allCredentials.filter(function (item){ return item.Id === id });

		if (matches.length < 1) {
			poop.reject(new Error("cred not found"));
			return;
		}

		poop.resolve(matches[0]);

		return poop;
	});
}
*/

function exposedCred(id) {
	return $.getJSON('/secrets/' + id + '/expose');
}

function credentialById(id) {
	return $.getJSON('/secrets/' + id);
}
