import { coerceToStructuredErrorResponse, isNotSignedInError, isSealedError } from 'backenderrors';
import { WarningAlert } from 'components/alerts';
import { Panel } from 'components/bootstrap';
import { Breadcrumb } from 'components/breadcrumbtrail';
import { CommandInlineForm } from 'components/CommandButton';
import { Loading } from 'components/loading';
import { DatabaseUnseal, SessionSignIn } from 'generated/commanddefinitions';
import { RootFolderId, RootFolderName } from 'generated/domain';
import { getFolder } from 'generated/restapi';
import DefaultLayout from 'layouts/DefaultLayout';
import * as React from 'react';
import { indexRoute } from 'routes';
import { shouldAlwaysSucceed, unrecognizedValue } from 'utils';

enum UnauthenticatedKind {
	Sealed, // while database is sealed, signing in is not possible
	AwaitingUsername,
	AwaitingPassword,
}

interface SignInPageProps {
	redirect: string;
}

interface SignInPageState {
	status?: UnauthenticatedKind;
	username: string;
}

export default class SignInPage extends React.Component<SignInPageProps, SignInPageState> {
	state: SignInPageState = {
		username: localStorage.getItem('signInLastUsername') || '',
	};
	private title = 'Sign in';

	componentDidMount() {
		shouldAlwaysSucceed(this.fetchData());
	}

	render() {
		const widget =
			this.state.status !== undefined ? this.widgetByStatus(this.state.status) : <Loading />;

		return (
			<DefaultLayout title={this.title} breadcrumbs={this.getBreadcrumbs()}>
				{widget}
			</DefaultLayout>
		);
	}

	private widgetByStatus(status: UnauthenticatedKind): JSX.Element {
		switch (status) {
			case UnauthenticatedKind.Sealed:
				return (
					<Panel heading="Unseal">
						<WarningAlert text="Database was sealed. Please unseal it." />

						<CommandInlineForm command={DatabaseUnseal()} />
					</Panel>
				);
			case UnauthenticatedKind.AwaitingUsername:
				return (
					<Panel heading="Username">
						<form
							onSubmit={() => {
								this.usernameEntered();
							}}>
							<label>
								Username:
								<input
									type="text"
									value={this.state.username}
									onChange={(e) => {
										this.setState({ username: e.target.value });
									}}
								/>
							</label>
							<input type="submit" value="Submit" />
						</form>
					</Panel>
				);
			case UnauthenticatedKind.AwaitingPassword:
				return (
					<Panel heading={'Username: ' + this.state.username}>
						<CommandInlineForm command={SessionSignIn(this.state.username)} />
					</Panel>
				);
			default:
				return unrecognizedValue(status);
		}
	}

	private usernameEntered() {
		// store, so next on next login we can pre-fill this
		localStorage.setItem('signInLastUsername', this.state.username);

		this.setState({ status: UnauthenticatedKind.AwaitingPassword });
	}

	private getBreadcrumbs(): Breadcrumb[] {
		return [
			{ url: indexRoute.buildUrl({}), title: RootFolderName },
			{ url: '', title: this.title },
		];
	}

	private async fetchData() {
		const status = await this.determineUnauthenticatedKind();

		if (status !== null) {
			this.setState({ status });
		} else {
			// signed in => redirect to where we wanted to go
			window.location.assign(this.props.redirect);
		}
	}

	private async determineUnauthenticatedKind(): Promise<UnauthenticatedKind | null> {
		try {
			// dummy request just to gauge problems status
			await getFolder(RootFolderId);

			return null;
		} catch (ex) {
			const ser = coerceToStructuredErrorResponse(ex);

			if (isSealedError(ser)) {
				return UnauthenticatedKind.Sealed;
			} else if (isNotSignedInError(ser)) {
				return this.state.username === ''
					? UnauthenticatedKind.AwaitingUsername
					: UnauthenticatedKind.AwaitingPassword;
			}

			throw ex; // some other error - shouldn't happen
		}
	}
}
