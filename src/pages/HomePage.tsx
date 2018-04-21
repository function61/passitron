import * as React from 'react';
import DefaultLayout from 'layouts/DefaultLayout';
import {FolderResponse} from 'model';
import {folderLink} from 'links';
import {getFolder, defaultErrorHandler} from 'repo';
import {CommandButton} from 'components/CommandButton';
import {createAccount, createFolder, renameFolder, moveFolder} from 'generated/commanddefinitions';
import {Breadcrumb} from 'components/breadcrumbtrail';
import {SecretListing} from 'components/SecretListing';

interface HomePageProps {
	folderId: string;
}

interface HomePageState {
	listing?: FolderResponse;
}

export default class HomePage extends React.Component<HomePageProps, HomePageState> {
	componentDidMount() {
		this.fetchData(this.props.folderId);
	}

	componentWillReceiveProps(nextProps: HomePageProps) {
		this.fetchData(nextProps.folderId);
	}

	render() {
		if (!this.state || !this.state.listing) {
			return <h1>loading</h1>;
		}

		const listing = this.state.listing;

		const breadcrumbs: Breadcrumb[] = [
			{ url: '', title: listing.Folder!.Name },
		];

		for (const parent of listing.ParentFolders) {
			breadcrumbs.unshift({ url: folderLink(parent.Id), title: parent.Name });
		}

		return <DefaultLayout breadcrumbs={breadcrumbs}>
			<SecretListing searchTerm="" listing={listing} />

			<CommandButton command={createAccount(this.props.folderId)}></CommandButton>
			<CommandButton command={createFolder(this.props.folderId)}></CommandButton>
			<CommandButton command={renameFolder(this.props.folderId, listing.Folder!.Name)}></CommandButton>
			<CommandButton command={moveFolder(this.props.folderId)}></CommandButton>
		</DefaultLayout>;
	}

	private fetchData(folderId: string) {
		getFolder(folderId).then((listing: FolderResponse) => {
			this.setState({ listing });
		}, defaultErrorHandler);
	}
}
