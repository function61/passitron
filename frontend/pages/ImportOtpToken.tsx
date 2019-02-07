import { CommandButton } from 'f61ui/component/CommandButton';
import { AccountAddOtpToken } from 'generated/commands_commands';
import * as QrCode from 'jsqrcode';
import { AppDefaultLayout } from 'layout/appdefaultlayout';
import * as React from 'react';

interface ImportOtpTokenProps {
	account: string;
}

interface ImportOtpTokenState {
	OtpProvisioningUrl?: string;
}

export default class ImportOtpToken extends React.Component<
	ImportOtpTokenProps,
	ImportOtpTokenState
> {
	private title = 'Import OTP token';

	render() {
		const breadcrumbs = [{ url: '', title: this.title }];

		const maybeSubmit =
			this.state && this.state.OtpProvisioningUrl ? (
				<CommandButton
					command={AccountAddOtpToken(this.props.account, this.state.OtpProvisioningUrl)}
				/>
			) : (
				<p>button will appear here</p>
			);

		return (
			<AppDefaultLayout title={this.title} breadcrumbs={breadcrumbs}>
				<h1>Import OTP token from QR code: {this.props.account}</h1>

				<input
					type="file"
					id="upload"
					onChange={(e) => {
						this.fileChange(e);
					}}
				/>

				{maybeSubmit}

				<h2>Or import manually</h2>

				<CommandButton command={AccountAddOtpToken(this.props.account, '')} />
			</AppDefaultLayout>
		);
	}

	fileChange(e: React.ChangeEvent<HTMLInputElement>) {
		if (!e.target.files) {
			return;
		}

		if (e.target.files.length !== 1) {
			throw new Error('Expecting one file');
		}

		const file = e.target.files[0];

		if (!/^image\//.test(file.type)) {
			throw new Error('Unsupported image type - must be image/*');
		}

		const qrReader = new QrCode();
		qrReader.callback = (err: Error, result: any) => {
			if (err) {
				alert('error reading QR code: ' + err.toString());
				return;
			}

			this.setState({
				OtpProvisioningUrl: result.result,
			});
		};

		const fileReader = new FileReader();
		fileReader.addEventListener(
			'load',
			() => {
				qrReader.decode(fileReader.result);
			},
			false,
		);
		fileReader.readAsDataURL(file);
	}
}
