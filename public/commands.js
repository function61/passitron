
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
			Password: {}
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
			Password: {},
			PasswordRepeat: {}
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
			MasterPassword: {}
		}
	},
	'ChangeDescriptionRequest': {
		fields: {
			Id: {},
			Description: {}
		}
	}
};

function inputDialog(opts) {
	var modal = $('<div class="modal fade" id="myModal" tabindex="-1" role="dialog">');

	var dialog = $('<div class="modal-dialog" role="document">').appendTo(modal);

	var content = $('<div class="modal-content">').appendTo(dialog);

	var header = $('<div class="modal-header">').appendTo(content);

	// for close, you'd use: <button type="button" class="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>

	$('<h4 class="modal-title"></h4>').text(opts.title).appendTo(header);

	var body = $('<div class="modal-body">').appendTo(content);

	var footer = $('<div class="modal-footer">').appendTo(content);

	$('<button type="button" class="btn btn-primary">Save</button>').appendTo(footer).click(function (){
		// TODO: totally remove from DOM
		$(modal).modal('hide');

		opts.ok(modal);
	});

	$(modal).modal('show')

	return { body: body };
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

		var type = 'text';

		if (type === 'text') {
			var input = $('<input type="text" class="form-control" />')
				.attr('name', field)
				.attr('id', id);

			if (opts.prefill && (field in opts.prefill)) {
				input.val(opts.prefill[ field ]);
			}

			input.appendTo(formGroup);
		} else {
			throw new Error('Unknown type: ' + type);
		}
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

	formCommand.appendTo(dlg.body);
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
