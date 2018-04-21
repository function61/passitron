import * as React from 'react';
import DefaultLayout from 'layouts/DefaultLayout';

interface ImportOtpTokenProps {
	account: string;
}

export default class ImportOtpToken extends React.Component<ImportOtpTokenProps, {}> {
	render() {
		const breadcrumbs = [
			{ url: '', title: 'Import OTP token' },
		];

		return <DefaultLayout breadcrumbs={breadcrumbs}>
			<h1>Import OTP token from QR code: {this.props.account}</h1>

			<input type="file" id="upload" />
		</DefaultLayout>;
	}
}
