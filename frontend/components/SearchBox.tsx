import * as React from 'react';
import {indexRoute, searchRoute} from 'routes';

interface SearchBoxProps {
	searchTerm?: string;
}

export class SearchBox extends React.Component<SearchBoxProps, {}> {
	render() {
		return <input
			type="text"
			style={{width: '250px'}}
			className="form-control"
			defaultValue={this.props.searchTerm}
			onKeyPress={(e) => { this.onSubmit(e); }}
			placeholder="Search .." />;
	}

	onSubmit(e: React.KeyboardEvent<HTMLInputElement>) {
		if (e.charCode !== 13) {
			return;
		}

		const searchTerm = e.currentTarget.value;

		if (searchTerm !== '') {
			document.location.hash = searchRoute.buildUrl({searchTerm});
		} else {
			document.location.hash = indexRoute.buildUrl({});
		}
	}
}
