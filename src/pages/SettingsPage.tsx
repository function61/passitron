import * as React from 'react';
import DefaultLayout from 'layouts/DefaultLayout';
import {Breadcrumb} from 'components/breadcrumbtrail';
import {CommandButton} from 'components/CommandButton';
import {rootFolderName} from 'model';
import {changeMasterPassword} from 'generated/commanddefinitions';
import {indexLink} from 'links';

export default class SettingsPage extends React.Component<{}, {}> {
	private getBreadcrumbs(): Breadcrumb[] {
		return [
			{url: indexLink(), title: rootFolderName},
			{url: '', title: 'settings'},
		];
	}

	render() {
		return <DefaultLayout breadcrumbs={this.getBreadcrumbs()}>
			<h1>Settings</h1>

			<button className="btn btn-default">unseal</button>
			<CommandButton command={changeMasterPassword()}></CommandButton>

			<h2>Export / import</h2>

			<button className="btn btn-default">write keepass</button>

		</DefaultLayout>;
	}
}
