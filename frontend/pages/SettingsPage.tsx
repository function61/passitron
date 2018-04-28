import * as React from 'react';
import DefaultLayout from 'layouts/DefaultLayout';
import {Breadcrumb} from 'components/breadcrumbtrail';
import {CommandButton} from 'components/CommandButton';
import {rootFolderName} from 'model';
import {DatabaseUnseal, DatabaseChangeMasterPassword, DatabaseExportToKeepass} from 'generated/commanddefinitions';
import {indexLink} from 'links';

export default class SettingsPage extends React.Component<{}, {}> {
	private getBreadcrumbs(): Breadcrumb[] {
		return [
			{url: indexLink(), title: rootFolderName},
			{url: '', title: 'settings'},
		];
	}

	render() {
		return <DefaultLayout title="Settings" breadcrumbs={this.getBreadcrumbs()}>
			<h1>Settings</h1>

			<CommandButton command={DatabaseUnseal()}></CommandButton>
			<CommandButton command={DatabaseChangeMasterPassword()}></CommandButton>

			<h2>Export / import</h2>

			<CommandButton command={DatabaseExportToKeepass()}></CommandButton>

		</DefaultLayout>;
	}
}