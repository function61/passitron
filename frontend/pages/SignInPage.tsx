import { U2fSigner } from 'components/U2F';
import { isNotSignedInError } from 'errors';
import { navigateTo } from 'f61ui/browserutils';
import { Button, Panel } from 'f61ui/component/bootstrap';
import { Breadcrumb } from 'f61ui/component/breadcrumbtrail';
import { CommandInlineForm } from 'f61ui/component/CommandButton';
import { CommandExecutor } from 'f61ui/component/commandpagelet';
import { Loading } from 'f61ui/component/loading';
import { Result } from 'f61ui/component/result';
import { coerceToStructuredErrorResponse, formatAnyError } from 'f61ui/errors';
import { shouldAlwaysSucceed, unrecognizedValue } from 'f61ui/utils';
import { SessionSignIn } from 'generated/apitypes_commands';
import { getFolder, getSignInChallenge } from 'generated/apitypes_endpoints';
import {
	ErrNeedU2fVerification,
	U2FChallengeBundle,
	U2FResponseBundle,
} from 'generated/apitypes_types';
import { RootFolderId, RootFolderName } from 'generated/domain_types';
import { AppDefaultLayout } from 'layout/appdefaultlayout';
import * as React from 'react';
import { indexRoute } from 'routes';

const storedUsernameLocalStorageKey = 'signInLastUsername';

enum UnauthenticatedKind {
	AwaitingUsername,
	AwaitingPassword,
	AwaitingU2fSignature,
	AwaitingLoggingIn,
}

interface SignInPageProps {
	redirect: string;
}

interface SignInPageState {
	status?: UnauthenticatedKind;
	username: string;
	signInChallenge: Result<U2FChallengeBundle>;
}

export default class SignInPage extends React.Component<SignInPageProps, SignInPageState> {
	state: SignInPageState = {
		username: localStorage.getItem(storedUsernameLocalStorageKey) || '',
		signInChallenge: new Result<U2FChallengeBundle>((x) => {
			this.setState({ signInChallenge: x });
		}),
	};
	private title = 'Sign in';
	private password = ''; // only set (and needed) if using U2F flow

	componentDidMount() {
		shouldAlwaysSucceed(this.fetchData());
	}

	render() {
		const widget =
			this.state.status !== undefined ? this.widgetByStatus(this.state.status) : <Loading />;

		return (
			<AppDefaultLayout title={this.title} breadcrumbs={this.getBreadcrumbs()}>
				<Panel heading={this.title}>{widget}</Panel>
			</AppDefaultLayout>
		);
	}

	private widgetByStatus(status: UnauthenticatedKind): React.ReactNode {
		switch (status) {
			case UnauthenticatedKind.AwaitingUsername:
				return (
					<form
						onSubmit={() => {
							this.rememberUsername();
						}}>
						<div className="form-group">
							<label>
								Username *
								<input
									type="text"
									className="form-control"
									value={this.state.username}
									autoFocus={true}
									onChange={(e) => {
										this.setState({ username: e.target.value });
									}}
								/>
							</label>
						</div>
						<input type="submit" value="Next" className="btn btn-primary" />
					</form>
				);
			case UnauthenticatedKind.AwaitingPassword:
				return (
					<div>
						<div className="form-group">
							<label>Username *</label>

							<p>
								{this.state.username}
								<span className="margin-left">
									<Button
										label="Change user"
										click={() => {
											this.forgetUsername();
										}}
									/>
								</span>
							</p>
						</div>
						<CommandInlineForm
							command={SessionSignIn(this.state.username, '', null, {
								error: (err, values) => {
									// only handle U2F verification
									if (err.error_code !== ErrNeedU2fVerification) {
										return false;
									}

									// need to store this because we'll try signing in again
									// after we get U2F signature
									this.password = values.Password;

									// this is a dirty hack
									const [userId, mac] = err.error_description.split(':');

									this.u2fFlowStart(userId, mac);

									return true; // error was handled
								},
							})}
						/>
					</div>
				);
			case UnauthenticatedKind.AwaitingU2fSignature:
				return this.state.signInChallenge.draw((signInChallenge) => (
					<U2fSigner
						challenge={signInChallenge}
						signed={(res) => {
							shouldAlwaysSucceed(this.u2fFlowSignIn(res));
						}}
					/>
				));
			case UnauthenticatedKind.AwaitingLoggingIn:
				return <Loading />;
			default:
				return unrecognizedValue(status);
		}
	}

	private rememberUsername() {
		if (this.state.username === '') {
			return;
		}

		// store, so next on next login we can pre-fill this
		localStorage.setItem(storedUsernameLocalStorageKey, this.state.username);

		this.setState({ status: UnauthenticatedKind.AwaitingPassword });
	}

	private forgetUsername() {
		localStorage.removeItem(storedUsernameLocalStorageKey);

		this.setState({ status: UnauthenticatedKind.AwaitingUsername });
	}

	private getBreadcrumbs(): Breadcrumb[] {
		return [
			{ url: indexRoute.buildUrl({}), title: RootFolderName },
			{ url: '', title: this.title },
		];
	}

	// signed in => redirect to where we wanted to go
	private successfullySignedInSoRedirect() {
		navigateTo(this.props.redirect);
	}

	private async fetchData() {
		const status = await this.determineUnauthenticatedKind();

		if (status !== null) {
			this.setState({ status });
		} else {
			this.successfullySignedInSoRedirect();
		}
	}

	private async determineUnauthenticatedKind(): Promise<UnauthenticatedKind | null> {
		try {
			// dummy request just to gauge status
			await getFolder(RootFolderId);

			return null;
		} catch (ex) {
			const err = coerceToStructuredErrorResponse(ex);

			if (isNotSignedInError(err)) {
				return this.state.username === ''
					? UnauthenticatedKind.AwaitingUsername
					: UnauthenticatedKind.AwaitingPassword;
			}

			throw ex; // some other error - shouldn't happen
		}
	}

	private u2fFlowStart(userId: string, mac: string) {
		this.setState({ status: UnauthenticatedKind.AwaitingU2fSignature });

		this.state.signInChallenge.load(() => getSignInChallenge(userId, mac));
	}

	// we got signature, now re-try signing in with signature in addition to username/pwd
	private async u2fFlowSignIn(signature: U2FResponseBundle) {
		this.setState({ status: UnauthenticatedKind.AwaitingLoggingIn });

		try {
			await new CommandExecutor(
				SessionSignIn(this.state.username, this.password, signature),
			).execute();

			this.successfullySignedInSoRedirect();
		} catch (err) {
			alert(formatAnyError(err));
		}
	}
}
