import {Breadcrumb} from 'components/breadcrumbtrail';
import {CommandLink} from 'components/CommandButton';
import {Dropdown} from 'components/dropdown';
import {Loading} from 'components/loading';
import {SecretListing} from 'components/SecretListing';
import {FolderResponse} from 'generated/apitypes';
import {AccountCreate, AccountCreateFolder, AccountMoveFolder, AccountRenameFolder} from 'generated/commanddefinitions';
import {defaultErrorHandler, getFolder} from 'generated/restapi';
import DefaultLayout from 'layouts/DefaultLayout';
import * as React from 'react';
import {folderRoute} from 'routes';

interface HomePageProps {
	folderId: string;
}

interface HomePageState {
	listing: FolderResponse;
}

export default class HomePage extends React.Component<HomePageProps, HomePageState> {
	componentDidMount() {
		this.fetchData(this.props.folderId);
	}

	componentWillReceiveProps(nextProps: HomePageProps) {
		this.fetchData(nextProps.folderId);
	}

	render() {
		if (!this.state) {
			return <Loading />;
		}

		const listing = this.state.listing;

		const breadcrumbs: Breadcrumb[] = [
			{ url: '', title: listing.Folder!.Name },
		];

		for (const parent of listing.ParentFolders) {
			breadcrumbs.unshift({ url: folderRoute.buildUrl({folderId: parent.Id}), title: parent.Name });
		}

		return <DefaultLayout title="Home" breadcrumbs={breadcrumbs}>
			<SecretListing listing={listing} />

			<Dropdown label="Actions">
				<CommandLink command={AccountCreate(this.props.folderId)}></CommandLink>
				<CommandLink command={AccountCreateFolder(this.props.folderId)}></CommandLink>
				<CommandLink command={AccountRenameFolder(this.props.folderId, listing.Folder!.Name)}></CommandLink>
				<CommandLink command={AccountMoveFolder(this.props.folderId)}></CommandLink>
			</Dropdown>
		</DefaultLayout>;
	}

	private fetchData(folderId: string) {
		getFolder(folderId).then((listing) => {
			this.setState({ listing });
		}, defaultErrorHandler);
	}
}
