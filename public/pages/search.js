
routes.search = function(args) {
	var search = args[1];

	allCreds(search).then(function(allCredentials){
		// $('<h1>loq</h1>').appendTo(document.body);

		var matches = allCredentials;

		credsWidget(matches, search).appendTo(cc());
	});
}

