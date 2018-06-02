import {Breadcrumb} from 'components/breadcrumbtrail';
import {CommandButton} from 'components/CommandButton';
import {DatabaseChangeMasterPassword, DatabaseExportToKeepass} from 'generated/commanddefinitions';
import {RootFolderName} from 'generated/domain';
import DefaultLayout from 'layouts/DefaultLayout';
import * as React from 'react';
import {indexRoute} from 'routes';

export default class SettingsPage extends React.Component<{}, {}> {
	private title = 'Settings';

	render() {
		return <DefaultLayout title={this.title} breadcrumbs={this.getBreadcrumbs()}>
			<h1>{this.title}</h1>

			<CommandButton command={DatabaseChangeMasterPassword()}></CommandButton>

			<h2>Export / import</h2>

			<CommandButton command={DatabaseExportToKeepass()}></CommandButton>

		</DefaultLayout>;
	}

	private getBreadcrumbs(): Breadcrumb[] {
		return [
			{url: indexRoute.buildUrl({}), title: RootFolderName},
			{url: '', title: this.title },
		];
	}
}
