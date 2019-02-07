import { defaultErrorHandler } from 'backenderrors';
import { SecretListing } from 'components/SecretListing';
import { Loading } from 'f61ui/component/loading';
import { getFolder } from 'generated/apitypes_endpoints';
import { FolderResponse } from 'generated/apitypes_types';
import { RootFolderId } from 'generated/domain_types';
import { AppDefaultLayout } from 'layout/appdefaultlayout';
import * as React from 'react';

interface SshKeysPageState {
	listing: FolderResponse;
}

export default class SshKeysPage extends React.Component<{}, SshKeysPageState> {
	private title = 'SSH keys';

	componentDidMount() {
		this.fetchData();
	}

	render() {
		if (!this.state) {
			return <Loading />;
		}

		const breadcrumbs = [{ url: '', title: this.title }];

		return (
			<AppDefaultLayout title={this.title} breadcrumbs={breadcrumbs}>
				<SecretListing searchTerm="" listing={this.state.listing} />
			</AppDefaultLayout>
		);
	}

	private fetchData() {
		getFolder(RootFolderId).then((listing) => {
			this.setState({ listing });
		}, defaultErrorHandler);
	}
}
