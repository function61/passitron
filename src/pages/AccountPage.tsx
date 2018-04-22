import * as React from 'react';
import {Account, Secret, SecretKind, Folder, FolderResponse} from 'model';
import clipboard from 'clipboard';
import {getAccount, getFolder, getSecrets} from 'repo';
import {Breadcrumb} from 'components/breadcrumbtrail';
import {CommandButton, CommandLink} from 'components/CommandButton';
import DefaultLayout from 'layouts/DefaultLayout';
import {folderLink, importOtpTokenLink} from 'links';
import {
	deleteAccount,
	addPassword,
	addSshKey,
	deleteSecret,
	changeUsername,
	changeDescription,
	renameAccount,
} from 'generated/commanddefinitions';
import {unrecognizedValue} from 'utils';

interface ShittyDropdownProps {
	children: JSX.Element[];
}

export class ShittyDropdown extends React.Component<ShittyDropdownProps, {}> {
	render() {
		const items = this.props.children.map((child, idx) => {
			return <li key={idx}>{child}</li>;
		});

		return <div className="btn-group">
			<button type="button" className="btn btn-default dropdown-toggle" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
				<span className="caret"></span>
			</button>
			<ul className="dropdown-menu">
				{ items }
			</ul>
		</div>;
	}
}

interface AccountPageProps {
	id: string;
}

interface AccountPageState {
	account: Account;
	secrets: Secret[];
	folderresponse: FolderResponse;
}

export default class AccountPage extends React.Component<AccountPageProps, AccountPageState> {
	// https://developmentarc.gitbooks.io/react-indepth/content/life_cycle/the_life_cycle_recap.html
	componentDidMount() {
		this.fetchData();
	}

	render() {
		if (!this.state || !this.state.account) {
			return <h1>loading</h1>;
		}

		const account = this.state.account;

		const secretRows = this.state.secrets.map((secret) => this.secretToRow(secret, account));

		const breadcrumbItems = this.getBreadcrumbItems();

		return <DefaultLayout breadcrumbs={breadcrumbItems}>
			<h1>
				{account.Title}
				&nbsp;
				<ShittyDropdown>
					<a href="#">Separated link</a>
					<a href="#">Separated link</a>
				</ShittyDropdown>
			</h1>

			<table className="table table-striped">
			<tbody>
				<tr>
					<td>
						Username
						<CommandLink command={changeUsername(account.Id, account.Username)} />
					</td>
					<td>{account.Username}</td>
					<td onClick={() => this.copyToClipboard(account.Username)}>📋</td>
				</tr>
				<tr>
					<td>Secrets</td>
					<td>
						<table className="table table-striped">
						<tbody>
							{secretRows}
						</tbody>
						</table>
					</td>
					<td></td>
				</tr>
				<tr>
					<td>
						Description
						<CommandLink command={changeDescription(account.Id, account.Description)} />
					</td>
					<td style={{fontFamily: 'monospace', whiteSpace: 'pre'}}>{account.Description}</td>
					<td></td>
				</tr>
			</tbody>
			</table>

			<CommandButton command={renameAccount(account.Id, account.Title)} />
			<CommandButton command={deleteAccount(account.Id)} />

			<CommandButton command={addSshKey(account.Id)} />
			<CommandButton command={addPassword(account.Id)} />

			<a href={importOtpTokenLink(account.Id)} className="btn btn-default">+ OTP token</a>

		</DefaultLayout>;
	}

	private secretToRow(secret: Secret, account: Account): JSX.Element {
			switch (secret.Kind) {
			case SecretKind.SshKey:			
				return <tr key={secret.Id}>
					<td>
						SSH public key
						<CommandLink command={deleteSecret(account.Id, secret.Id)} />
					</td>
					<td>{secret.SshPublicKeyAuthorized}</td>
					<td></td>
				</tr>;
			case SecretKind.Password:
				return <tr key={secret.Id}>
					<td>
						Password
						<CommandLink command={deleteSecret(account.Id, secret.Id)} />
					</td>
					<td>{secret.Password}</td>
					<td onClick={() => this.copyToClipboard(secret.Password)}>📋</td>
				</tr>;
			case SecretKind.OtpToken:
				return <tr key={secret.Id}>
					<td>
						OTP
						<CommandLink command={deleteSecret(account.Id, secret.Id)} />
					</td>
					<td>{secret.OtpProof}</td>
					<td onClick={() => this.copyToClipboard(secret.OtpProof)}></td>
				</tr>;
			default:
				return unrecognizedValue(secret.Kind);
			}
	}

	private fetchData() {
		const account = getAccount(this.props.id);

		const folder = account.then((acc) => getFolder(acc.FolderId));

		const secrets = account.then((acc) => getSecrets(acc.Id));

		Promise.all([account, folder, secrets]).then(([account, folderresponse, secrets]) => {
			this.setState({
				account,
				folderresponse,
				secrets,
			});
		});
	}

	private getBreadcrumbItems(): Breadcrumb[] {
		const breadcrumbItems: Breadcrumb[] = [
			{ url: '', title: this.state.account.Title },
		];

		function unshiftFolderToBreadcrumb(fld: Folder) {
			breadcrumbItems.unshift({
				url: folderLink(fld.Id),
				title: fld.Name,
			});
		}

		unshiftFolderToBreadcrumb(this.state.folderresponse.Folder!);
		this.state.folderresponse.ParentFolders.forEach((fld) => {
			unshiftFolderToBreadcrumb(fld);
		});

		return breadcrumbItems;
	}

	private copyToClipboard(secretToClipboard: string) {
		if (!clipboard(secretToClipboard)) {
			alert('failed to copy to clipboard');
		}
	}
}
