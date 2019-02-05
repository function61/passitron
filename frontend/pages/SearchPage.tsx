import { defaultErrorHandler } from 'backenderrors';
import { SecretListing } from 'components/SecretListing';
import { Breadcrumb } from 'f61ui/components/breadcrumbtrail';
import { Loading } from 'f61ui/components/loading';
import { FolderResponse } from 'generated/apitypes';
import { RootFolderName } from 'generated/domain';
import { search } from 'generated/restapi';
import { AppDefaultLayout } from 'layout/appdefaultlayout';
import * as React from 'react';
import { indexRoute } from 'routes';

interface SearchPageProps {
	searchTerm: string;
}

interface SearchPageState {
	searchResult: FolderResponse;
}

export default class SearchPage extends React.Component<SearchPageProps, SearchPageState> {
	componentDidMount() {
		this.fetchData();
	}

	render() {
		if (!this.state) {
			return <Loading />;
		}

		const breadcrumbs: Breadcrumb[] = [
			{ url: indexRoute.buildUrl({}), title: RootFolderName },
			{ url: '', title: `Search: ${this.props.searchTerm}` },
		];

		return (
			<AppDefaultLayout title="Search" breadcrumbs={breadcrumbs}>
				<SecretListing
					searchTerm={this.props.searchTerm}
					listing={this.state.searchResult}
				/>
			</AppDefaultLayout>
		);
	}

	private fetchData() {
		search(this.props.searchTerm).then((searchResult) => {
			this.setState({ searchResult });
		}, defaultErrorHandler);
	}
}
