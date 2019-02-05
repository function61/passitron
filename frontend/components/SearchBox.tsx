import { defaultErrorHandler } from 'backenderrors';
import { navigateTo } from 'f61ui/browserutils';
import { search } from 'generated/restapi';
import * as React from 'react';
import * as Autocomplete from 'react-autocomplete';
import { indexRoute, searchRoute } from 'routes';
import { accountRoute, folderRoute } from 'routes';

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
	private queuedQuery: undefined | string;

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
							navigateTo(searchRoute.buildUrl({ searchTerm }));
						} else {
							navigateTo(indexRoute.buildUrl({}));
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
		if (this.hasInflight) {
			this.queuedQuery = term;
			return;
		}

		this.hasInflight = true;

		const scheduleNextOk = () => {
			setTimeout(() => {
				this.hasInflight = false;
				const qq = this.queuedQuery;
				this.queuedQuery = undefined;
				if (qq !== undefined) {
					this.searchtermChanged(qq);
				}
			}, 1000);
		};

		search(term).then(
			(resp) => {
				const folderMatches: AutocompleteItem[] = resp.SubFolders.map((item) => ({
					label: item.Name,
					url: folderRoute.buildUrl({ folderId: item.Id }),
				}));

				const accountMatches: AutocompleteItem[] = resp.Accounts.map((item) => ({
					label: item.Title,
					url: accountRoute.buildUrl({ id: item.Id }),
				}));

				this.setState({ items: folderMatches.concat(accountMatches) });

				scheduleNextOk();
			},
			(err) => {
				scheduleNextOk();
				defaultErrorHandler(err);
			},
		);
	}
}
