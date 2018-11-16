import {defaultErrorHandler} from 'backenderrors';
import {Panel} from 'components/bootstrap';
import {Breadcrumb} from 'components/breadcrumbtrail';
import {CommandButton} from 'components/CommandButton';
import {Loading} from 'components/loading';
import {Timestamp} from 'components/timestamp';
import {RegisterResponse, U2FEnrolledToken} from 'generated/apitypes';
import {DatabaseChangeMasterPassword, DatabaseExportToKeepass, UserRegisterU2FToken} from 'generated/commanddefinitions';
import {RootFolderName} from 'generated/domain';
import {u2fEnrolledTokens, u2fEnrollmentChallenge} from 'generated/restapi';
import DefaultLayout from 'layouts/DefaultLayout';
import * as React from 'react';
import {indexRoute} from 'routes';
import {isU2FError, U2FStdRegisteredKey, U2FStdRegisterRequest, U2FStdRegisterResponse} from 'u2ftypes';

interface SettingsPageState {
	u2fregistrationrequest?: string;
	enrolledTokens?: U2FEnrolledToken[];
}

export default class SettingsPage extends React.Component<{}, SettingsPageState> {
	state: SettingsPageState = {};
	private title = 'Settings';

	componentDidMount() {
		this.fetchData();
	}

	render() {
		const enrollOrFinish = this.state.u2fregistrationrequest ?
			<CommandButton command={UserRegisterU2FToken(this.state.u2fregistrationrequest)}></CommandButton> :
			<p><a className="btn btn-default" onClick={() => {this.startTokenEnrollment(); }}>Enroll token</a></p>;

		const enrolledTokensList = this.state.enrolledTokens ?
			<table className="table">
			<thead>
			<tr>
				<th>Name</th>
				<th>Type</th>
				<th>EnrolledAt</th>
			</tr>
			</thead>
			<tbody>{this.state.enrolledTokens.map((token) =>
			<tr key={token.EnrolledAt}>
				<td>{token.Name}</td>
				<td>{token.Version}</td>
				<td><Timestamp ts={token.EnrolledAt} /></td>
			</tr>)}
			</tbody>
			</table> :
			<Loading />;

		return <DefaultLayout title={this.title} breadcrumbs={this.getBreadcrumbs()}>
			<Panel heading="Actions">
				<div><CommandButton command={DatabaseChangeMasterPassword()}></CommandButton></div>

				<div className="margin-top"><CommandButton command={DatabaseExportToKeepass()}></CommandButton></div>
			</Panel>

			<Panel heading="U2F tokens">
				<h3>Enrolled tokens</h3>

				{enrolledTokensList}

				{enrollOrFinish}
			</Panel>
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

	private fetchData() {
		u2fEnrolledTokens().then((enrolledTokens) => {
			this.setState({ enrolledTokens });
		}, defaultErrorHandler);
	}
}
