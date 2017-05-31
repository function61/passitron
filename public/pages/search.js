
routes.search = function(args) {
	var search = args[1];

	rest_search(search).then(function(entries){
		breadcrumbWidget([ { label: 'Search', href: '' }]).appendTo(cc());

		credsWidget([], entries, search).appendTo(cc());
	});
}

