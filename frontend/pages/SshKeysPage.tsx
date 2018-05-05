import * as React from 'react';
import DefaultLayout from 'layouts/DefaultLayout';
import {FolderResponse} from 'model';
import {RootFolderId} from 'generated/domain';
import {getFolder, defaultErrorHandler} from 'repo';
import {SecretListing} from 'components/SecretListing';

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
