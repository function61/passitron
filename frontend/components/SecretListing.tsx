import { CommandLink } from 'components/CommandButton';
import { Dropdown } from 'components/dropdown';
import { SearchBox } from 'components/SearchBox';
import { FolderResponse } from 'generated/apitypes';
import {
	AccountDeleteFolder,
	AccountMove,
	AccountMoveFolder,
	AccountRenameFolder,
} from 'generated/commanddefinitions';
import * as React from 'react';
import { accountRoute, folderRoute } from 'routes';

interface SecretListingProps {
	searchTerm?: string;
	listing: FolderResponse;
}

export class SecretListing extends React.Component<SecretListingProps, {}> {
	render() {
		const folderRows = this.props.listing.SubFolders.map((folder) => {
			return (
				<tr key={folder.Id}>
					<td>
						<span className="glyphicon glyphicon-folder-open" />
					</td>
					<td>
						<a href={folderRoute.buildUrl({ folderId: folder.Id })}>{folder.Name}</a>
					</td>
					<td />
					<td>
						<Dropdown>
							<CommandLink command={AccountRenameFolder(folder.Id, folder.Name)} />
							<CommandLink command={AccountMoveFolder(folder.Id)} />
							<CommandLink command={AccountDeleteFolder(folder.Id)} />
						</Dropdown>
					</td>
				</tr>
			);
		});

		const accountRows = this.props.listing.Accounts.map((account) => {
			return (
				<tr key={account.Id}>
					<td />
					<td>
						<a href={accountRoute.buildUrl({ id: account.Id })}>{account.Title}</a>
					</td>
					<td>{account.Username}</td>
					<td>
						<Dropdown>
							<CommandLink command={AccountMove(account.Id)} />
						</Dropdown>
					</td>
				</tr>
			);
		});

		return (
			<div>
				<table className="table table-striped">
					<thead>
						<tr>
							<th />
							<th>
								Title
								<br />
								<SearchBox searchTerm={this.props.searchTerm} />
							</th>
							<th>Username</th>
							<th />
						</tr>
					</thead>
					<tbody>
						{folderRows}
						{accountRows}
					</tbody>
				</table>
			</div>
		);
	}
}
