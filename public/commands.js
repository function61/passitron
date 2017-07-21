
var commands = {
	'FolderCreateRequest': {
		fields: {
			ParentId: {},
			Name: {}
		}
	},
	'SecretCreateRequest': {
		fields: {
			FolderId: {},
			Title: {},
			Username: {},
			Password: {
				type: 'password'
			}
		}
	},
	'RenameSecretRequest': {
		fields: {
			Id: {},
			Title: {}
		}
	},
	'ChangeUsernameRequest': {
		fields: {
			Id: {},
			Username: {}
		}
	},
	'ChangePasswordRequest': {
		fields: {
			Id: {},
			Password: {
				type: 'password'
			},
			PasswordRepeat: {
				type: 'password'
			}
		}
	},
	'SetSshKeyRequest': {
		fields: {
			Id: {},
			SshPrivateKey: {
				type: 'textarea'
			}
		}
	},
	'SetOtpTokenRequest': {
		fields: {
			Id: {},
			OtpProvisioningUrl: {}
		}
	},
	'WriteKeepassRequest': {
		fields: { }
	},
	'DeleteSecretRequest': {
		fields: {
			Id: {}
		}
	},
	'RenameFolderRequest': {
		fields: {
			Id: {},
			Name: {}
		}
	},
	'MoveFolderRequest': {
		fields: {
			Id: {},
			ParentId: {}
		}
	},
	'UnsealRequest': {
		fields: {
			MasterPassword: {
				type: 'password'
			}
		}
	},
	'ChangeMasterPasswordRequest': {
		fields: {
			NewMasterPassword: {
				type: 'password'
			},
			NewMasterPasswordRepeat: {
				type: 'password'
			}
		}
	},
	'ChangeDescriptionRequest': {
		fields: {
			Id: {},
			Description: {
				type: 'textarea'
			}
		}
	}
};

function inputDialog(opts) {
	var modal = $('<div class="modal fade" tabindex="-1" role="dialog">');

	var dialog = $('<div class="modal-dialog" role="document">').appendTo(modal);

	var content = $('<div class="modal-content">').appendTo(dialog);

	var header = $('<div class="modal-header">').appendTo(content);

	// for close, you'd use: <button type="button" class="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>

	$('<h4 class="modal-title"></h4>').text(opts.title).appendTo(header);

	var body = $('<div class="modal-body">').appendTo(content);

	var footer = $('<div class="modal-footer">').appendTo(content);

	function close() {
		// TODO: totally remove from DOM
		$(modal).modal('hide');

		opts.ok(modal);
	}

	$('<button type="button" class="btn btn-primary">Save</button>').appendTo(footer).click(close);

	$(modal).modal('show');

	return { body: body, close: close, modal: modal };
}

var runningId = 0;

function nextRunningId() {
	runningId++;

	return runningId.toString();
}

function createFormForCommand(cmdSpec, opts) {
	var form = $('<form></form>');

	for (var field in cmdSpec.fields) {
		var fieldSpec = cmdSpec.fields[ field ];

		var formGroup = $('<div class="form-group"></div>').appendTo(form);

		var id = 'cmdui' + nextRunningId();

		$('<label></label>')
			.attr('for', id)
			.text(field).appendTo(formGroup);

		var type = fieldSpec.type || 'text';

		var input;
		if (type === 'text') {
			input = $('<input type="text" class="form-control" />')
				.attr('name', field)
				.attr('id', id);
		} else if (type === 'password') {
			input = $('<input type="password" class="form-control" />')
				.attr('name', field)
				.attr('id', id);
		} else if (type === 'textarea') {
			input = $('<textarea rows="7" class="form-control" />')
				.attr('name', field)
				.attr('id', id);
		} else {
			throw new Error('Unknown type: ' + type);
		}

		if (opts.prefill && (field in opts.prefill)) {
			input.val(opts.prefill[ field ]);
		}

		input.appendTo(formGroup);
	}

	return form;
	/*
	<form>
  <div class="form-group">
    <label for="exampleInputEmail1">Email address</label>
    <input type="email" class="form-control" id="exampleInputEmail1" placeholder="Email">
  </div>
  */
}

function invokeCommand(cmd, opts) {
	opts = opts || {};

	var cmdSpec = commands[ cmd ];
	var formCommand = createFormForCommand(cmdSpec, opts);

	var dlg = inputDialog({
		title: cmd,
		ok: function (){
			var values = {};

			dlg.body.find(':input').each(function (){
				values[ this.name ] = this.value;
			});

			$.ajax({
				type: 'POST',
				url: '/command/' + cmd,
				data: JSON.stringify(values),
				success: function(data) {
					softReload();
				},
				error: function (xhr) {
					alert('xhr error: ' + xhr.responseText);
				},
				contentType: "application/json"
			});
		}
	});

	formCommand.on('submit', function (){
		dlg.close();

		return false; // abort
	});

	formCommand.appendTo(dlg.body);

	// https://stackoverflow.com/a/11634933
	dlg.modal.on('shown.bs.modal', function (){
		var firstInput = formCommand.find(':input')[0];

		if (firstInput) {
			firstInput.select();
		}
	});
}

function invokeCommandFromClickEvent() {
	var node = this;

	var cmd = node.getAttribute('data-cmd');
	var opts = JSON.parse(node.getAttribute('data-cmd-opts') || '{}');

	invokeCommand(cmd, opts);
}

function initCommandArchitecture() {
	// attach listener delegate
	$(document).on('click', '.command', invokeCommandFromClickEvent);
}

function attachCommand(dom, spec) {
	dom.addClass('command').attr('data-cmd', spec.cmd);

	var opts = {};

	if (spec.prefill) {
		opts.prefill = spec.prefill;
	}

	if (Object.keys(opts).length > 0) {
		dom.attr('data-cmd-opts', JSON.stringify(opts));
	}
};
