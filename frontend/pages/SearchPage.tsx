import * as React from 'react';
import {Account, FolderResponse} from 'model';
import {RootFolderName} from 'generated/domain';
import {indexRoute} from 'routes';
import {searchAccounts, defaultErrorHandler} from 'repo';
import DefaultLayout from 'layouts/DefaultLayout';
import {SecretListing} from 'components/SecretListing';
import {Breadcrumb} from 'components/breadcrumbtrail';

interface SearchPageProps {
	searchTerm: string;
}

interface SearchPageState {
	matches: Account[];
}

export default class SearchPage extends React.Component<SearchPageProps, SearchPageState> {
	componentDidMount() {
		this.fetchData();
	}

	render() {
		if (!this.state) {
			return <h1>loading</h1>;
		}

		// adapt for reuse for direct use of <SecretListing> component
		const dummyResult: FolderResponse = {
			Folder: null,
			SubFolders: [],
			ParentFolders: [],
			Accounts: this.state.matches,
		};

		const breadcrumbs: Breadcrumb[] = [
			{ url: indexRoute.buildUrl({}), title: RootFolderName },
			{ url: '', title: `Search: ${this.props.searchTerm}` },
		];

		return <DefaultLayout title="Search" breadcrumbs={breadcrumbs}>
			<SecretListing searchTerm={this.props.searchTerm} listing={dummyResult} />
		</DefaultLayout>;
	}

	private fetchData() {
		searchAccounts(this.props.searchTerm).then((matches: Account[]) => {
			this.setState({ matches });
		}, defaultErrorHandler);
	}
}
