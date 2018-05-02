import * as React from 'react';
import DefaultLayout from 'layouts/DefaultLayout';
import {Breadcrumb} from 'components/breadcrumbtrail';
import {CommandButton} from 'components/CommandButton';
import {rootFolderName} from 'model';
import {DatabaseUnseal, DatabaseChangeMasterPassword, DatabaseExportToKeepass} from 'generated/commanddefinitions';
import {indexRoute} from 'routes';

export default class SettingsPage extends React.Component<{}, {}> {
	private title = 'Settings';

	private getBreadcrumbs(): Breadcrumb[] {
		return [
			{url: indexRoute.buildUrl({}), title: rootFolderName},
			{url: '', title: this.title },
		];
	}

	render() {
		return <DefaultLayout title={this.title} breadcrumbs={this.getBreadcrumbs()}>
			<h1>Settings</h1>

			<CommandButton command={DatabaseUnseal()}></CommandButton>
			<CommandButton command={DatabaseChangeMasterPassword()}></CommandButton>

			<h2>Export / import</h2>

			<CommandButton command={DatabaseExportToKeepass()}></CommandButton>

		</DefaultLayout>;
	}
}
