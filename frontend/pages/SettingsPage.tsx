import { Panel } from 'f61ui/component/bootstrap';
import { Breadcrumb } from 'f61ui/component/breadcrumbtrail';
import { CommandButton, CommandLink } from 'f61ui/component/CommandButton';
import { Dropdown } from 'f61ui/component/dropdown';
import { Loading } from 'f61ui/component/loading';
import { Timestamp } from 'f61ui/component/timestamp';
import { defaultErrorHandler } from 'f61ui/errors';
import { shouldAlwaysSucceed } from 'f61ui/utils';
import { u2fEnrolledTokens, u2fEnrollmentChallenge, userList } from 'generated/apitypes_endpoints';
import { RegisterResponse, U2FEnrolledToken, User } from 'generated/apitypes_types';
import {
	DatabaseChangeMasterPassword,
	DatabaseExportToKeepass,
	SessionSignOut,
	UserAddAccessToken,
	UserChangePassword,
	UserCreate,
	UserRegisterU2FToken,
} from 'generated/commands_commands';
import { RootFolderName } from 'generated/domain_types';
import { AppDefaultLayout } from 'layout/appdefaultlayout';
import * as React from 'react';
import { indexRoute } from 'routes';
import {
	isU2FError,
	U2FStdRegisteredKey,
	U2FStdRegisterRequest,
	U2FStdRegisterResponse,
} from 'u2ftypes';

interface SettingsPageState {
	u2fregistrationrequest?: string;
	enrolledTokens?: U2FEnrolledToken[];
	users?: User[];
}

export default class SettingsPage extends React.Component<{}, SettingsPageState> {
	state: SettingsPageState = {};
	private title = 'Settings';

	componentDidMount() {
		shouldAlwaysSucceed(this.fetchData());
	}

	render() {
		const enrollOrFinish = this.state.u2fregistrationrequest ? (
			<CommandButton command={UserRegisterU2FToken(this.state.u2fregistrationrequest)} />
		) : (
			<p>
				<a
					className="btn btn-default"
					onClick={() => {
						this.startTokenEnrollment();
					}}>
					Enroll token
				</a>
			</p>
		);

		return (
			<AppDefaultLayout title={this.title} breadcrumbs={this.getBreadcrumbs()}>
				<div className="row">
					<div className="col-md-4">
						<Panel heading="Actions">
							<div>
								<CommandButton command={DatabaseChangeMasterPassword()} />
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

							{enrollOrFinish}
						</Panel>
					</div>
				</div>
			</AppDefaultLayout>
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

			const reqs: U2FStdRegisterRequest[] = res.RegisterRequest.RegisterRequests.map(
				(item) => {
					return {
						version: item.Version,
						challenge: item.Challenge,
					};
				},
			);

			const keys: U2FStdRegisteredKey[] = res.RegisterRequest.RegisteredKeys.map((item) => {
				return {
					version: item.Version,
					keyHandle: item.KeyHandle,
					appId: item.AppID,
				};
			});

			u2f.register(res.RegisterRequest.AppID, reqs, keys, u2fRegisterCallback, 30);
		}, defaultErrorHandler);
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
