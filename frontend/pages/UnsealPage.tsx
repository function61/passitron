import {coerceToStructuredErrorResponse, defaultErrorHandler, isSealedError} from 'backenderrors';
import {WarningAlert} from 'components/alerts';
import {Breadcrumb} from 'components/breadcrumbtrail';
import {CommandInlineForm} from 'components/CommandButton';
import {Loading} from 'components/loading';
import {DatabaseUnseal} from 'generated/commanddefinitions';
import {RootFolderId, RootFolderName} from 'generated/domain';
import {getFolder} from 'generated/restapi';
import DefaultLayout from 'layouts/DefaultLayout';
import * as React from 'react';
import {indexRoute} from 'routes';

interface UnsealPageProps {
	redirect: string;
}

interface UnsealPageState {
	unsealed: boolean;
}

export default class UnsealPage extends React.Component<UnsealPageProps, UnsealPageState> {
	private title = 'Unseal';

	componentDidMount() {
		this.fetchData();
	}

	render() {
		let content = <Loading />;

		if (this.state) {
			if (!this.state.unsealed) {
				content = <div>
					<WarningAlert text="Database was sealed. Please unseal it." />

					<CommandInlineForm command={DatabaseUnseal()} />
				</div>;
			} else {
				// content = <SuccessAlert text="Unsealed successfully." />;
				throw new Error('This should not happen anymore');
			}
		}

		return <DefaultLayout title={this.title} breadcrumbs={this.getBreadcrumbs()}>
			<h1>{this.title}</h1>

			{content}

		</DefaultLayout>;
	}

	private getBreadcrumbs(): Breadcrumb[] {
		return [
			{url: indexRoute.buildUrl({}), title: RootFolderName},
			{url: '', title: this.title},
		];
	}

	private fetchData() {
		// dummy request just to identify unsealed status
		getFolder(RootFolderId).then(() => {
			window.location.assign(this.props.redirect);
		}, (err) => {
			if (isSealedError(coerceToStructuredErrorResponse(err))) {
				this.setState({ unsealed: false });
				return;
			}

			// some other error
			defaultErrorHandler(err);
		});
	}
}
