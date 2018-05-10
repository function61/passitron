import {FolderResponse} from 'generated/apitypes';
import * as React from 'react';
import {credviewRoute, folderRoute, indexRoute, searchRoute} from 'routes';

interface SecretListingProps {
	searchTerm: string;
	listing: FolderResponse;
}

export class SecretListing extends React.Component<SecretListingProps, {}> {
	render() {
		const folderRows = this.props.listing.SubFolders.map((folder) => {
			return <tr key={folder.Id}>
				<td><span className="glyphicon glyphicon-folder-open"></span></td>
				<td><a href={folderRoute.buildUrl({folderId: folder.Id})}>{folder.Name}</a></td>
				<td></td>
			</tr>;
		});

		const accountRows = this.props.listing.Accounts.map((account) => {
			return <tr key={account.Id}>
				<td></td>
				<td><a href={credviewRoute.buildUrl({id: account.Id})}>{account.Title}</a></td>
				<td>{account.Username}</td>
			</tr>;
		});

		return <div>
			<table className="table table-striped">
			<thead>
				<tr>
					<th></th>
					<th>
						Title<br />
						<input type="text" style={{width: '250px'}} className="form-control" defaultValue={this.props.searchTerm} onKeyPress={(e) => this.onSubmit(e)} placeholder="Search .." />
					</th>
					<th>Username</th>
				</tr>
			</thead>
			<tbody>
			{folderRows}
			{accountRows}
			</tbody>
			</table>
		</div>;
	}

	// onSubmit(e: KeyboardEvent<HTMLInputElement>) {
	onSubmit(e: any) {
		if (e.charCode !== 13) {
			return;
		}

		const searchTerm = e.target.value;

		if (searchTerm !== '') {
			document.location.hash = searchRoute.buildUrl({searchTerm});
		} else {
			document.location.hash = indexRoute.buildUrl({});
		}
	}
}
