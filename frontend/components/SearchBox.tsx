import { navigateTo } from 'f61ui/browserutils';
import { defaultErrorHandler } from 'f61ui/errors';
import { shouldAlwaysSucceed } from 'f61ui/utils';
import { search } from 'generated/apitypes_endpoints';
import * as React from 'react';
import * as Autocomplete from 'react-autocomplete';
import { accountUrl, folderUrl, indexUrl, searchUrl } from 'generated/apitypes_uiroutes';

interface SearchBoxProps {
	searchTerm?: string; // initial
}

interface SearchBoxState {
	searchTerm?: string;
	items: AutocompleteItem[];
	hasInflight: boolean;
}

interface AutocompleteItem {
	label: string;
	url: string;
}

export class SearchBox extends React.Component<SearchBoxProps, SearchBoxState> {
	state: SearchBoxState = { hasInflight: false, items: [] };
	private hasInflight = false;
	private beginSearchTimeout?: ReturnType<typeof setTimeout>;
	private queuedQuery = '';

	render() {
		return (
			<Autocomplete
				inputProps={{
					onKeyPress: (e: any) => {
						// https://github.com/reactjs/react-autocomplete/issues/338
						if (e.key !== 'Enter') {
							return;
						}

						// => user hit enter with non-suggested term => go to search results
						const searchTerm = e.target.value;

						if (searchTerm !== '') {
							navigateTo(searchUrl({ q: searchTerm }));
						} else {
							navigateTo(indexUrl());
						}
					},
				}}
				getItemValue={(item: AutocompleteItem) => item.url}
				items={this.state.items}
				renderItem={(item: AutocompleteItem, isHighlighted: boolean) => (
					<div
						style={{ background: isHighlighted ? 'lightgray' : 'white' }}
						key={item.url}>
						{item.label}
					</div>
				)}
				value={this.state.searchTerm}
				onChange={(e: any) => {
					this.setState({ searchTerm: e.target.value, items: [] });
					this.searchtermChanged(e.target.value);
				}}
				onSelect={(url: string) => {
					navigateTo(url);
				}}
			/>
		);
	}

	private searchtermChanged(term: string) {
		this.queuedQuery = term;

		if (this.beginSearchTimeout || this.hasInflight) {
			return;
		}

		this.beginSearchTimeout = setTimeout(() => {
			this.beginSearchTimeout = undefined;

			shouldAlwaysSucceed(this.search(this.queuedQuery));
		}, 500);
	}

	private async search(term: string) {
		this.queuedQuery = '';

		this.hasInflight = true;

		try {
			const searchResult = await search(term);

			const folderMatches: AutocompleteItem[] = searchResult.SubFolders.map((item) => ({
				label: item.Name,
				url: folderUrl({ id: item.Id }),
			}));

			const accountMatches: AutocompleteItem[] = searchResult.Accounts.map((item) => {
				const label = item.Username ? `${item.Title} (${item.Username})` : item.Title;
				return {
					label,
					url: accountUrl({ id: item.Id }),
				};
			});

			this.setState({ items: folderMatches.concat(accountMatches) });
		} catch (err) {
			defaultErrorHandler(err);
		}

		this.hasInflight = false;

		// while we were fetching data from server, user wanted another query?
		if (this.queuedQuery) {
			shouldAlwaysSucceed(this.search(this.queuedQuery));
		}
	}
}
