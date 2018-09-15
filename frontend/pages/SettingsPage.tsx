import {Breadcrumb} from 'components/breadcrumbtrail';
import {CommandButton} from 'components/CommandButton';
import {RegisterResponse} from 'generated/apitypes';
import {DatabaseChangeMasterPassword, DatabaseExportToKeepass, UserRegisterU2FToken} from 'generated/commanddefinitions';
import {RootFolderName} from 'generated/domain';
import {defaultErrorHandler, u2fEnrollmentChallenge} from 'generated/restapi';
import DefaultLayout from 'layouts/DefaultLayout';
import * as React from 'react';
import {indexRoute} from 'routes';
import {isU2FError, U2FStdRegisteredKey, U2FStdRegisterRequest, U2FStdRegisterResponse} from 'u2ftypes';

interface SettingsPageState {
	u2fregistrationrequest?: string;
}

export default class SettingsPage extends React.Component<{}, SettingsPageState> {
	state: SettingsPageState = {};
	private title = 'Settings';

	render() {
		const enrollOrFinish = this.state.u2fregistrationrequest ?
			<CommandButton command={UserRegisterU2FToken(this.state.u2fregistrationrequest)}></CommandButton> :
			<p><a className="btn btn-default" onClick={() => {this.startTokenEnrollment(); }}>Enroll token</a></p>;

		return <DefaultLayout title={this.title} breadcrumbs={this.getBreadcrumbs()}>
			<h1>{this.title}</h1>

			<CommandButton command={DatabaseChangeMasterPassword()}></CommandButton>

			<h2>Export / import</h2>

			<CommandButton command={DatabaseExportToKeepass()}></CommandButton>

			<h2>U2F token</h2>

			{enrollOrFinish}
		</DefaultLayout>;
	}

	private startTokenEnrollment() {
		u2fEnrollmentChallenge().then((res) => {
			const challenge = res.Challenge;

			const u2fRegisterCallback = (regResponse: U2FStdRegisterResponse) => {
				if (isU2FError(regResponse)) {
					return;
				}

				const enrollmentRequest: RegisterResponse = {
					Challenge: challenge,
					RegisterResponse: {
						RegistrationData: regResponse.registrationData,
						Version: regResponse.version,
						ClientData: regResponse.clientData,
					},
				};

				const enrollmentRequestAsJson = JSON.stringify(enrollmentRequest);

				this.setState({ u2fregistrationrequest: enrollmentRequestAsJson });
			};

			const reqs: U2FStdRegisterRequest[] = res.RegisterRequest.RegisterRequests.map((item) => {
				return {
					version: item.Version,
					challenge: item.Challenge,
				};
			});

			const keys: U2FStdRegisteredKey[] = res.RegisterRequest.RegisteredKeys.map((item) => {
				return {
					version: item.Version,
					keyHandle: item.KeyHandle,
					appId: item.AppID,
				};
			});

			u2f.register(
				res.RegisterRequest.AppID,
				reqs,
				keys,
				u2fRegisterCallback,
				30);
		}, defaultErrorHandler);
	}

	private getBreadcrumbs(): Breadcrumb[] {
		return [
			{url: indexRoute.buildUrl({}), title: RootFolderName},
			{url: '', title: this.title },
		];
	}
}
