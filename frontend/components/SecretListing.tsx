import * as React from 'react';
import {FolderResponse} from 'model';
import {searchLink, folderLink, indexLink, secretLink} from 'links';

interface SecretListingProps {
	searchTerm: string;
	listing: FolderResponse;
}

export class SecretListing extends React.Component<SecretListingProps, {}> {
	render() {
		const folderRows = this.props.listing.SubFolders.map((folder) => {
			return <tr key={folder.Id}>
				<td><span className="glyphicon glyphicon-folder-open"></span></td>
				<td><a href={folderLink(folder.Id)}>{folder.Name}</a></td>
				<td></td>
			</tr>;
		});

		const accountRows = this.props.listing.Accounts.map((account) => {
			return <tr key={account.Id}>
				<td></td>
				<td><a href={secretLink(account.Id)}>{account.Title}</a></td>
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
						<input type="text" style={{width: '250px'}} className="form-control" defaultValue={this.props.searchTerm} onKeyPress={e => this.onSubmit(e)} placeholder="Search .." />
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
			document.location.hash = searchLink(searchTerm);
		} else {
			document.location.hash = indexLink();
		}
	}
}
