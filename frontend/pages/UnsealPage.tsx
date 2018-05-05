import {WarningAlert} from 'components/alerts';
import {Breadcrumb} from 'components/breadcrumbtrail';
import {CommandButton} from 'components/CommandButton';
import {DatabaseUnseal} from 'generated/commanddefinitions';
import {RootFolderName} from 'generated/domain';
import DefaultLayout from 'layouts/DefaultLayout';
import * as React from 'react';
import {indexRoute} from 'routes';

export default class UnsealPage extends React.Component<{}, {}> {
	private title = 'Unseal';

	render() {
		return <DefaultLayout title={this.title} breadcrumbs={this.getBreadcrumbs()}>
			<h1>Unseal</h1>

			<WarningAlert text="Database was sealed. Please unseal it." />

			<CommandButton command={DatabaseUnseal()}></CommandButton>

		</DefaultLayout>;
	}

	private getBreadcrumbs(): Breadcrumb[] {
		return [
			{url: indexRoute.buildUrl({}), title: RootFolderName},
			{url: '', title: this.title},
		];
	}
}
