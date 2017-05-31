
routes.credview = function (args) {
	var id = args[1];

	credentialById(id).then(function (cred){
		var titleHeading = $('<h1></h1>')
			.text(cred.Title)
			.appendTo(cc());

		attachCommand(titleHeading, {
			cmd: 'RenameSecretRequest',
			prefill: {
				Id: id,
				Title: cred.Title
			} });

		var detailsTable = createTable();

		var tr;

		tr = detailsTable.tr();
		detailsTable.td(tr).text('Username');
		detailsTable.td(tr).attr('id', 'username').text(cred.Username);
		detailsTable.td(tr).attr('data-clipboard-target', '#username').text('📋');

		var statusTr = detailsTable.tr();
		detailsTable.td(statusTr).text('Password');
		detailsTable.td(statusTr).text('.. requesting authorization ..');

		exposedCred(cred.Id).then(function (exposeResult){
			statusTr.remove();

			tr = detailsTable.tr();
			detailsTable.td(tr).text('Password');
			detailsTable.td(tr).attr('id', 'pwd').text(exposeResult.Password);
			detailsTable.td(tr).attr('data-clipboard-target', '#pwd').text('📋');
		});

		var descriptionTr = detailsTable.tr();
		var descriptionHeadingTd = detailsTable.td(descriptionTr).text('Description');
		detailsTable.td(descriptionTr).text(cred.Description);

		attachCommand(descriptionHeadingTd, {
			cmd: 'ChangeDescriptionRequest',
			prefill: {
				Id: id,
				Description: cred.Description
			} });


		detailsTable.table.appendTo(cc());

		var secretDeleteBtn = $('<button class="btn btn-default"></button>')
			.text('Delete')
			.appendTo(cc());

		attachCommand(secretDeleteBtn, {
			cmd: 'DeleteSecretRequest',
			prefill: {
				Id: id
			} });

		var changePasswordBtn = $('<button class="btn btn-default"></button>')
			.text('Change pwd')
			.appendTo(cc());

		attachCommand(changePasswordBtn, {
			cmd: 'ChangePasswordRequest',
			prefill: {
				Id: id
			} });
	});
}
