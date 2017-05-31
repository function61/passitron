
routes.settings = function () {
	$('<h1>Settings</h1>').appendTo(cc());

	var writeKeepassBtn = $('<button class="btn btn-default"></button>')
		.text('write keepass')
		.appendTo(cc());

	attachCommand(writeKeepassBtn, { cmd: 'WriteKeepassRequest' } );
};
