
routes.sshkeys = function(args) {
	$('<h1>SSH keys</h1>').appendTo(cc());

	rest_sshkeys().then(function (entries){
		var tbl = createTable();

		tbl.th().text('Entry');
		tbl.th().text('Public key');

		for (var i = 0; i < entries.length; ++i) {
			var tr = tbl.tr();

			var a = $('<a></a>').appendTo(tbl.td(tr));
			a.text(entries[i].Title);
			a.attr('href', linkTo([ 'credview', entries[i].Id ]));

			// tbl.td(tr).text(entries[i].SshPublicKeyAuthorized.substr(0, 64) + '...');
			tbl.td(tr).text('?');
		}

		tbl.table.appendTo(cc());
	}, restDefaultErrorHandler);
}
