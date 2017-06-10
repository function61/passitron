
routes.settings = function () {
	$('<h1>Settings</h1>').appendTo(cc());

	var unsealBtn = $('<button class="btn btn-default"></button>')
		.text('Unseal')
		.appendTo(cc());

	attachCommand(unsealBtn, { cmd: 'UnsealRequest' } );

	$('<h3>Export / import</h3>').appendTo(cc());

	var writeKeepassBtn = $('<button class="btn btn-default"></button>')
		.text('write keepass')
		.appendTo(cc());

	attachCommand(writeKeepassBtn, { cmd: 'WriteKeepassRequest' } );
};
