
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

		var usernameTr = detailsTable.tr();
		var usernameHeading = detailsTable.td(usernameTr).text('Username');
		detailsTable.td(usernameTr).attr('id', 'username').text(cred.Username);
		detailsTable.td(usernameTr).attr('data-clipboard-target', '#username').text('📋');

		attachCommand(usernameHeading, {
			cmd: 'ChangeUsernameRequest',
			prefill: {
				Id: id,
				Username: cred.Username
			} });

		var pwdTr = detailsTable.tr();
		var pwdHeading = detailsTable
			.td(pwdTr)
			.text('Password');
		var pwdTd = detailsTable
			.td(pwdTr)
			.attr('id', 'pwd')
			.text('.. requesting authorization ..');
		detailsTable
			.td(pwdTr)
			.attr('data-clipboard-target', '#pwd')
			.text('📋');

		attachCommand(pwdHeading, {
			cmd: 'ChangePasswordRequest',
			prefill: {
				Id: id
			} });

		var tfaTr = detailsTable.tr();
		detailsTable.td(tfaTr).text('TFA proof');
		var tfaProofTd = detailsTable.td(tfaTr).attr('id', 'tfaproof');
		detailsTable.td(tfaTr).attr('data-clipboard-target', '#tfaproof').text('📋');

		tfaTr.hide();

		exposedCred(cred.Id).then(function (exposeResult){
			pwdTd.text(exposeResult.Password);

			if (exposeResult.OtpProof !== "") {
				tfaProofTd.text(exposeResult.OtpProof);
				tfaTr.show();
			}
		});

		var descriptionTr = detailsTable.tr();
		var descriptionHeadingTd = detailsTable.td(descriptionTr).text('Description');
		detailsTable
			.td(descriptionTr)
			.css('font-family', 'monospace')
			.css('white-space', 'pre')
			.text(cred.Description);

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

		$('<a class="btn btn-default">Attach OTP token</a>')
			.attr('href', linkTo([ 'importotptoken', id ]))
			.appendTo(cc());
	});
}
