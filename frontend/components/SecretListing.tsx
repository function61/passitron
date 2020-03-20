import { SearchBox } from 'components/SearchBox';
import { CommandLink } from 'f61ui/component/CommandButton';
import { Dropdown } from 'f61ui/component/dropdown';
import {
	AccountDeleteFolder,
	AccountMove,
	AccountMoveFolder,
	AccountRenameFolder,
} from 'generated/apitypes_commands';
import { FolderResponse } from 'generated/apitypes_types';
import * as React from 'react';
import { accountUrl, folderUrl } from 'generated/apitypes_uiroutes';

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
						<a href={folderUrl({ id: folder.Id })}>{folder.Name}</a>
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
						<a href={accountUrl({ id: account.Id })}>{account.Title}</a>
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
