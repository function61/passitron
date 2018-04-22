import * as React from 'react';
import DefaultLayout from 'layouts/DefaultLayout';
import {Breadcrumb} from 'components/breadcrumbtrail';
import {CommandButton} from 'components/CommandButton';
import {WarningAlert} from 'components/alerts';
import {rootFolderName} from 'model';
import {unseal} from 'generated/commanddefinitions';
import {indexLink} from 'links';

export default class UnsealPage extends React.Component<{}, {}> {
	private getBreadcrumbs(): Breadcrumb[] {
		return [
			{url: indexLink(), title: rootFolderName},
			{url: '', title: 'unseal'},
		];
	}

	render() {
		return <DefaultLayout breadcrumbs={this.getBreadcrumbs()}>
			<h1>Unseal</h1>

			<WarningAlert text="Database was sealed. Please unseal it." />

			<CommandButton command={unseal()}></CommandButton>

		</DefaultLayout>;
	}
}
