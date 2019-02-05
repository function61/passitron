import { defaultErrorHandler } from 'backenderrors';
import { SecretListing } from 'components/SecretListing';
import { Breadcrumb } from 'f61ui/component/breadcrumbtrail';
import { CommandLink } from 'f61ui/component/CommandButton';
import { Dropdown } from 'f61ui/component/dropdown';
import { Loading } from 'f61ui/component/loading';
import { FolderResponse } from 'generated/apitypes';
import { AccountCreate, AccountCreateFolder } from 'generated/commanddefinitions';
import { getFolder } from 'generated/restapi';
import { AppDefaultLayout } from 'layout/appdefaultlayout';
import * as React from 'react';
import { folderRoute } from 'routes';

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

		const breadcrumbs: Breadcrumb[] = [{ url: '', title: listing.Folder!.Name }];

		for (const parent of listing.ParentFolders) {
			breadcrumbs.unshift({
				url: folderRoute.buildUrl({ folderId: parent.Id }),
				title: parent.Name,
			});
		}

		return (
			<AppDefaultLayout title="Home" breadcrumbs={breadcrumbs}>
				<SecretListing listing={listing} />

				<Dropdown label="New ..">
					<CommandLink command={AccountCreate(this.props.folderId)} />
					<CommandLink command={AccountCreateFolder(this.props.folderId)} />
				</Dropdown>
			</AppDefaultLayout>
		);
	}

	private fetchData(folderId: string) {
		getFolder(folderId).then((listing) => {
			this.setState({ listing });
		}, defaultErrorHandler);
	}
}
