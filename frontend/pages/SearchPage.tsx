import {Breadcrumb} from 'components/breadcrumbtrail';
import {Loading} from 'components/loading';
import {SecretListing} from 'components/SecretListing';
import {FolderResponse} from 'generated/apitypes';
import {RootFolderName} from 'generated/domain';
import {defaultErrorHandler, search} from 'generated/restapi';
import DefaultLayout from 'layouts/DefaultLayout';
import * as React from 'react';
import {indexRoute} from 'routes';

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

		return <DefaultLayout title="Search" breadcrumbs={breadcrumbs}>
			<SecretListing searchTerm={this.props.searchTerm} listing={this.state.searchResult} />
		</DefaultLayout>;
	}

	private fetchData() {
		search(this.props.searchTerm).then((searchResult) => {
			this.setState({ searchResult });
		}, defaultErrorHandler);
	}
}
