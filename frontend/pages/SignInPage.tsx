import {coerceToStructuredErrorResponse, isNotSignedInError, isSealedError} from 'backenderrors';
import {WarningAlert} from 'components/alerts';
import {Panel} from 'components/bootstrap';
import {Breadcrumb} from 'components/breadcrumbtrail';
import {CommandInlineForm} from 'components/CommandButton';
import {Loading} from 'components/loading';
import {DatabaseUnseal, SessionSignIn} from 'generated/commanddefinitions';
import {RootFolderId, RootFolderName} from 'generated/domain';
import {getFolder} from 'generated/restapi';
import DefaultLayout from 'layouts/DefaultLayout';
import * as React from 'react';
import {indexRoute} from 'routes';
import {shouldAlwaysSucceed, unrecognizedValue} from 'utils';

enum UnauthenticatedKind {
	Sealed, // while database is sealed, signing in is not possible
	NotSignedIn,
}

interface SignInPageProps {
	redirect: string;
}

interface SignInPageState {
	status: UnauthenticatedKind;
}

export default class SignInPage extends React.Component<SignInPageProps, SignInPageState> {
	private title = 'Sign in';

	componentDidMount() {
		shouldAlwaysSucceed(this.fetchData());
	}

	render() {
		let content = <Loading />;

		if (this.state) {
			content = this.widgetByStatus(this.state.status);
		}

		return <DefaultLayout title={this.title} breadcrumbs={this.getBreadcrumbs()}>
			{content}
		</DefaultLayout>;
	}

	private widgetByStatus(status: UnauthenticatedKind): JSX.Element {
		switch (status) {
		case UnauthenticatedKind.Sealed:
			return <Panel heading="Unseal">
				<WarningAlert text="Database was sealed. Please unseal it." />

				<CommandInlineForm command={DatabaseUnseal()} />
			</Panel>;
		case UnauthenticatedKind.NotSignedIn:
			return <Panel heading="Sign in">
				<WarningAlert text="You need to sign in." />

				<CommandInlineForm command={SessionSignIn()} />
			</Panel>;
		default:
			return unrecognizedValue(status);
		}
	}

	private getBreadcrumbs(): Breadcrumb[] {
		return [
			{url: indexRoute.buildUrl({}), title: RootFolderName},
			{url: '', title: this.title},
		];
	}

	private async fetchData() {
		const status = await this.determineUnauthenticatedKind();

		if (status !== null) {
			this.setState({ status });
		} else { // signed in => redirect to where we wanted to go
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
				return UnauthenticatedKind.NotSignedIn;
			}

			throw ex; // some other error - shouldn't happen
		}
	}
}
