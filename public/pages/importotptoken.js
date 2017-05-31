
routes.importotptoken = function(args) {
	var credId = args[1];

	$('<h1>Import OTP token from QR code</h1>').appendTo(cc());

	var fileInput = $('<input type="file" id="upload">').appendTo(cc());

	var qrReader = new QrCode();
	qrReader.callback = function(err, result) {
		if (err) {
			alert('error reading QR code');
			console.log(err);
			return;
		}

		invokeCommand('SetOtpTokenRequest', {
			prefill: {
				Id: credId,
				OtpProvisioningUrl: result.result
			}
		});
	}

	fileInput[0].addEventListener('change', function() {
		for (var i = 0; i < this.files.length; i++) {
			var file = this.files[i];
			if (!/^image\//.test(file.type)) {
				throw new Error('Unsupported image type - must be image/*');
			}

			var fileReader = new FileReader();
			fileReader.addEventListener('load', function() {
				qrReader.decode(fileReader.result);
			}, false);
			fileReader.readAsDataURL(file);
		}
	}, false);

};
