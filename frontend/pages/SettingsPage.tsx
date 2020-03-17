import { DangerAlert, InfoAlert } from 'f61ui/component/alerts';
import { Button, Panel } from 'f61ui/component/bootstrap';
import { Breadcrumb } from 'f61ui/component/breadcrumbtrail';
import { CommandButton, CommandInlineForm, CommandLink } from 'f61ui/component/CommandButton';
import { Dropdown } from 'f61ui/component/dropdown';
import { Loading } from 'f61ui/component/loading';
import { Timestamp } from 'f61ui/component/timestamp';
import { defaultErrorHandler, formatAnyError } from 'f61ui/errors';
import { shouldAlwaysSucceed } from 'f61ui/utils';
import {
	DatabaseExportToKeepass,
	SessionSignOut,
	UserAddAccessToken,
	UserChangeDecryptionKeyPassword,
	UserChangePassword,
	UserCreate,
	UserRegisterU2FToken,
} from 'generated/apitypes_commands';
import { u2fEnrolledTokens, u2fEnrollmentChallenge, userList } from 'generated/apitypes_endpoints';
import { RegisterResponse, U2FEnrolledToken, User } from 'generated/apitypes_types';
import { RootFolderName } from 'generated/domain_types';
import { AppDefaultLayout } from 'layout/appdefaultlayout';
import * as React from 'react';
import { indexRoute } from 'routes';
import { isU2FError, u2fErrorMsg, U2FStdRegisterResponse } from 'u2ftypes';

interface SettingsPageState {
	u2fregistrationrequest?: string;
	enrollmentInProgress: boolean;
	enrolledTokens?: U2FEnrolledToken[];
	enrollmentError?: string;
	users?: User[];
}

export default class SettingsPage extends React.Component<{}, SettingsPageState> {
	state: SettingsPageState = {
		enrollmentInProgress: false,
	};
	private title = 'Settings';

	componentDidMount() {
		shouldAlwaysSucceed(this.fetchData());
	}

	render() {
		return (
			<AppDefaultLayout title={this.title} breadcrumbs={this.getBreadcrumbs()}>
				<div className="row">
					<div className="col-md-4">
						<Panel heading="Actions">
							<div>
								<CommandButton command={UserChangeDecryptionKeyPassword()} />
							</div>

							<div className="margin-top">
								<CommandButton command={DatabaseExportToKeepass()} />
							</div>

							<div className="margin-top">
								<CommandButton command={SessionSignOut()} />
							</div>
						</Panel>
					</div>
					<div className="col-md-8">
						<Panel heading="Users">{this.renderUsers()}</Panel>

						<Panel heading="U2F tokens">
							<h3>Enrolled tokens</h3>

							{this.renderEnrolledTokens()}

							{this.u2fEnrollmentUi()}
						</Panel>
					</div>
				</div>
			</AppDefaultLayout>
		);
	}

	// shows:
	// - enrollment start button
	// - error
	// - progress
	// - finish
	private u2fEnrollmentUi(): React.ReactNode {
		if (this.state.enrollmentInProgress) {
			return (
				<div>
					<InfoAlert>Please swipe your U2F token now.</InfoAlert>
					<Loading />
				</div>
			);
		}

		if (this.state.u2fregistrationrequest) {
			return (
				<CommandInlineForm
					command={UserRegisterU2FToken(this.state.u2fregistrationrequest)}
				/>
			);
		}

		return (
			<p>
				{this.state.enrollmentError && (
					<DangerAlert>{this.state.enrollmentError}</DangerAlert>
				)}
				<Button
					label="Enroll token"
					click={() => {
						shouldAlwaysSucceed(this.startTokenEnrollment());
					}}
				/>
			</p>
		);
	}

	private renderUsers() {
		return this.state.users ? (
			<div>
				<table className="table">
					<thead>
						<tr>
							<th>Id</th>
							<th>Username</th>
							<th>Created</th>
							<th>Password last changed</th>
							<th />
						</tr>
					</thead>
					<tbody>
						{this.state.users.map((user) => (
							<tr key={user.Id}>
								<td>{user.Id}</td>
								<td>{user.Username}</td>
								<td>
									<Timestamp ts={user.Created} />
								</td>
								<td>
									<Timestamp ts={user.PasswordLastChanged} />
								</td>
								<td>
									<Dropdown>
										<CommandLink command={UserChangePassword(user.Id)} />
										<CommandLink command={UserAddAccessToken(user.Id)} />
									</Dropdown>
								</td>
							</tr>
						))}
					</tbody>
				</table>

				<CommandButton command={UserCreate()} />
			</div>
		) : (
			<Loading />
		);
	}

	private renderEnrolledTokens() {
		return this.state.enrolledTokens ? (
			<table className="table">
				<thead>
					<tr>
						<th>Name</th>
						<th>Type</th>
						<th>EnrolledAt</th>
					</tr>
				</thead>
				<tbody>
					{this.state.enrolledTokens.map((token) => (
						<tr key={token.EnrolledAt}>
							<td>{token.Name}</td>
							<td>{token.Version}</td>
							<td>
								<Timestamp ts={token.EnrolledAt} />
							</td>
						</tr>
					))}
				</tbody>
			</table>
		) : (
			<Loading />
		);
	}

	private async startTokenEnrollment() {
		this.setState({ enrollmentInProgress: true, enrollmentError: '' });

		try {
			await this.startTokenEnrollmentInternal();
		} catch (err) {
			this.setState({ enrollmentError: formatAnyError(err) });
		}

		this.setState({ enrollmentInProgress: false });
	}

	private async startTokenEnrollmentInternal() {
		const res = await u2fEnrollmentChallenge();

		const enrollmentRequest = await new Promise<RegisterResponse>((resolve, reject) => {
			u2f.register(
				res.RegisterRequest.AppID,
				res.RegisterRequest.RegisterRequests.map((item) => {
					return {
						version: item.Version,
						challenge: item.Challenge,
					};
				}),
				res.RegisterRequest.RegisteredKeys.map((item) => {
					return {
						version: item.Version,
						keyHandle: item.KeyHandle,
						appId: item.AppID,
					};
				}),
				(regResponse: U2FStdRegisterResponse) => {
					if (isU2FError(regResponse)) {
						reject(u2fErrorMsg(regResponse));
						return;
					}

					resolve({
						Challenge: res.Challenge,
						RegisterResponse: {
							RegistrationData: regResponse.registrationData,
							Version: regResponse.version,
							ClientData: regResponse.clientData,
						},
					});
				},
				30,
			);
		});

		this.setState({ u2fregistrationrequest: JSON.stringify(enrollmentRequest) });
	}

	private getBreadcrumbs(): Breadcrumb[] {
		return [
			{ url: indexRoute.buildUrl({}), title: RootFolderName },
			{ url: '', title: this.title },
		];
	}

	private async fetchData() {
		const fetchEnrolledTokens = async () => {
			const enrolledTokens = await u2fEnrolledTokens();
			this.setState({ enrolledTokens });
		};

		const fetchUsers = async () => {
			const users = await userList();
			this.setState({ users });
		};

		try {
			// does in parallel
			await Promise.all([fetchEnrolledTokens(), fetchUsers()]);
		} catch (ex) {
			defaultErrorHandler(ex);
		}
	}
}
