import { CommandButton } from 'f61ui/component/CommandButton';
import { AccountAddOtpToken } from 'generated/commands_commands';
import jsQR from 'jsqr';
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
		if (!e.target.files || e.target.files.length === 0) {
			return;
		}

		if (e.target.files.length !== 1) {
			throw new Error('Expecting one file');
		}

		const file = e.target.files[0];

		if (!/^image\//.test(file.type)) {
			throw new Error('Unsupported image type - must be image/*');
		}

		imageFromFile(file, (img: HTMLImageElement) => {
			const idata = imageDataFromImage(img);

			const scanResult = jsQR(idata.data, idata.width, idata.height);
			if (!scanResult) {
				alert('error reading QR code');
				return;
			}

			this.setState({
				OtpProvisioningUrl: scanResult.data,
			});
		});
	}
}

function imageDataFromImage(img: HTMLImageElement): ImageData {
	const tempCanvas = document.createElement('canvas');
	// these have to be set explicitly via direct width and height, and not via .style
	tempCanvas.width = img.width;
	tempCanvas.height = img.height;

	const canvasCtx = tempCanvas.getContext('2d');
	if (!canvasCtx) {
		throw new Error('canvas 2d rendering context not available');
	}

	// drawing on the canvas seems to work (at least in Chrome) when it's not in the
	// document, which is convenient b/c we don't have to do cleanup
	canvasCtx.drawImage(img, 0, 0);

	return canvasCtx.getImageData(0, 0, img.width, img.height);
}

function imageFromFile(file: File, load: (img: HTMLImageElement) => void) {
	const img = new Image();
	img.onload = () => {
		URL.revokeObjectURL(img.src);

		load(img);
	};
	img.src = URL.createObjectURL(file);
}
