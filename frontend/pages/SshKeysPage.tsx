import {SecretListing} from 'components/SecretListing';
import {FolderResponse} from 'generated/apitypes';
import {RootFolderId} from 'generated/domain';
import DefaultLayout from 'layouts/DefaultLayout';
import * as React from 'react';
import {defaultErrorHandler, getFolder} from 'repo';

interface SshKeysPageState {
	listing: FolderResponse;
}

export default class SshKeysPage extends React.Component<{}, SshKeysPageState> {
	private title = 'SSH keys';

	componentDidMount() {
		this.fetchData();
	}

	render() {
		if (!this.state || !this.state.listing) {
			return <h1>loading</h1>;
		}

		const breadcrumbs = [
			{ url: '', title: this.title },
		];

		return <DefaultLayout title={this.title} breadcrumbs={breadcrumbs}>
			<SecretListing searchTerm="" listing={this.state.listing} />
		</DefaultLayout>;
	}

	private fetchData() {
		getFolder(RootFolderId).then((listing) => {
			this.setState({ listing });
		}, defaultErrorHandler);
	}
}
