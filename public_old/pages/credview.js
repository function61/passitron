
var _randomDomId_counter = 0;

function randomDomId(prefix) {
	return prefix + (++_randomDomId_counter);
}

routes.credview = function (args) {
	var accountId = args[1];

	rest_credentialById(accountId).then(function (cred){
		var titleHeading = $('<h1></h1>')
			.text(cred.Title)
			.attr('title', 'created: ' + cred.Created)
			.appendTo(cc());

		attachCommand(titleHeading, {
			cmd: 'RenameSecretRequest',
			prefill: {
				Id: accountId,
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
				Id: accountId,
				Username: cred.Username
			} });

		var secretsTr = detailsTable.tr();
		var secretsHeading = detailsTable
			.td(secretsTr)
			.text('Secrets');

		var secretsTd = detailsTable
			.td(secretsTr)
			.attr('colspan', 2);

		var secretsTable = createTable();
		secretsTable.table.appendTo(secretsTd);

		rest_exposedCred(cred.Id).then(function (secrets){
			for (var i = 0; i < secrets.length; ++i) {
				var secret = secrets[i];

				// for clipboard.js references
				var secretDomId = randomDomId('secret');

				var secretTr = secretsTable.tr();
				var secretHeadingTd;

				if (secret.Kind === 'password') {
					secretHeadingTd = secretsTable
						.td(secretTr)
						.text('Password');
					secretsTable
						.td(secretTr)
						.attr('id', secretDomId)
						.attr('title', 'last changed: ' + secret.Created)
						.text(secret.Password);
					secretsTable
						.td(secretTr)
						.attr('data-clipboard-target', '#' + secretDomId)
						.text('📋');

				} else if (secret.Kind === 'otp_token') {
					secretHeadingTd = secretsTable
						.td(secretTr)
						.text('OTP');

					secretsTable
						.td(secretTr)
						.attr('id', secretDomId)
						.attr('title', 'last changed: ' + secret.Created)
						.text(secret.OtpProof);
					secretsTable
						.td(secretTr)
						.attr('data-clipboard-target', '#' + secretDomId)
						.text('📋');
				} else if (secret.Kind === 'ssh_key') {
					secretHeadingTd = secretsTable
						.td(secretTr)
						.text('SSH public key');

					secretsTable
						.td(secretTr)
						.attr('colspan', 2)
						.attr('title', 'last changed: ' + secret.Created)
						.text(secret.SshPublicKeyAuthorized);
				} else {
					throw new Error("Unknown secret kind: " + secret.Kind);
				}

				attachCommand(secretHeadingTd, {
					cmd: 'DeleteSecretRequest',
					prefill: {
						Account: accountId,
						Secret: secret.Id
					} });
			}
		}, restDefaultErrorHandler);

		var addSshKeyBtn = $('<button class="btn btn-default"></button>')
			.text('+ SSH key')
			.appendTo(secretsTd);

		attachCommand(addSshKeyBtn, {
			cmd: 'SetSshKeyRequest',
			prefill: {
				Id: accountId
			} });

		var addPasswordBtn = $('<button class="btn btn-default"></button>')
			.text('+ Password')
			.appendTo(secretsTd);

		attachCommand(addPasswordBtn, {
			cmd: 'ChangePasswordRequest',
			prefill: {
				Id: accountId
			} });

		$('<a class="btn btn-default">+ OTP token</a>')
			.attr('href', linkTo([ 'importotptoken', accountId ]))
			.appendTo(secretsTd);

		var descriptionTr = detailsTable.tr();
		var descriptionHeadingTd = detailsTable.td(descriptionTr).text('Description');
		detailsTable
			.td(descriptionTr)
			.attr('colspan', 2)
			.css('font-family', 'monospace')
			.css('white-space', 'pre')
			.text(cred.Description);

		attachCommand(descriptionHeadingTd, {
			cmd: 'ChangeDescriptionRequest',
			prefill: {
				Id: accountId,
				Description: cred.Description
			} });


		detailsTable.table.appendTo(cc());

		var secretDeleteBtn = $('<button class="btn btn-default"></button>')
			.text('Delete')
			.appendTo(cc());

		attachCommand(secretDeleteBtn, {
			cmd: 'DeleteAccountRequest',
			prefill: {
				Id: accountId
			} });
	}, restDefaultErrorHandler);
}
