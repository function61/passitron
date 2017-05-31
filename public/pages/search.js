
routes.search = function(args) {
	var search = args[1];

	searchWidget(search).appendTo(cc());

	allCreds(search).then(function(allCredentials){
		// $('<h1>loq</h1>').appendTo(document.body);

		var matches = allCredentials;

		credsWidget(matches).appendTo(cc());
	});
}

