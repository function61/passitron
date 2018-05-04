import * as React from 'react';
import {Account, Secret, SecretKind, Folder, FolderResponse} from 'model';
import clipboard from 'clipboard';
import {getAccount, getFolder, getSecrets, defaultErrorHandler} from 'repo';
import {Breadcrumb} from 'components/breadcrumbtrail';
import {SecretReveal} from 'components/secretreveal';
import {CommandButton, CommandLink} from 'components/CommandButton';
import DefaultLayout from 'layouts/DefaultLayout';
import {folderRoute, importotptokenRoute} from 'routes';
import {
	AccountDelete,
	AccountAddPassword,
	AccountAddSshKey,
	AccountDeleteSecret,
	AccountChangeUsername,
	AccountAddKeylist,
	AccountChangeDescription,
	AccountRename,
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

		return <DefaultLayout title={account.Title} breadcrumbs={breadcrumbItems}>
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
						<CommandLink command={AccountChangeUsername(account.Id, account.Username)} />
					</td>
					<td>{account.Username}</td>
					<td onClick={() => this.copyToClipboard(account.Username)} className="fauxlink">📋</td>
				</tr>
				{secretRows}
				<tr>
					<td>
						Description
						<CommandLink command={AccountChangeDescription(account.Id, account.Description)} />
					</td>
					<td style={{fontFamily: 'monospace', whiteSpace: 'pre'}}>{account.Description}</td>
					<td></td>
				</tr>
			</tbody>
			</table>

			<CommandButton command={AccountRename(account.Id, account.Title)} />
			<CommandButton command={AccountDelete(account.Id)} />

			<CommandButton command={AccountAddSshKey(account.Id)} />
			<CommandButton command={AccountAddKeylist(account.Id)} />
			<CommandButton command={AccountAddPassword(account.Id)} />

			<a href={importotptokenRoute.buildUrl({account: account.Id})} className="btn btn-default">+ OTP token</a>

		</DefaultLayout>;
	}

	private secretToRow(secret: Secret, account: Account): JSX.Element {
			switch (secret.Kind) {
			case SecretKind.SshKey:			
				return <tr key={secret.Id}>
					<td>
						SSH public key
						<CommandLink type="remove" command={AccountDeleteSecret(account.Id, secret.Id)} />
					</td>
					<td>{secret.SshPublicKeyAuthorized}</td>
					<td></td>
				</tr>;
			case SecretKind.Password:
				return <tr key={secret.Id}>
					<td>
						Password
						<CommandLink type="remove" command={AccountDeleteSecret(account.Id, secret.Id)} />
					</td>
					<td><SecretReveal secret={secret.Password} /></td>
					<td onClick={() => this.copyToClipboard(secret.Password)} className="fauxlink">📋</td>
				</tr>;
			case SecretKind.OtpToken:
				return <tr key={secret.Id}>
					<td>
						OTP
						<CommandLink type="remove" command={AccountDeleteSecret(account.Id, secret.Id)} />
					</td>
					<td>{secret.OtpProof}</td>
					<td onClick={() => this.copyToClipboard(secret.OtpProof)} className="fauxlink">📋</td>
				</tr>;
			case SecretKind.Keylist:
				const keyRows = secret.KeylistKeys.map((item) => <tr key={item.Key}>
					<th>{item.Key}</th>
					<td>{item.Value}</td>
				</tr>);

				return <tr key={secret.Id}>
					<td>
						Keylist
						<CommandLink type="remove" command={AccountDeleteSecret(account.Id, secret.Id)} />
					</td>
					<td colSpan={2}>{secret.Title}
						<table className="table table-striped">
						<tbody>{keyRows}</tbody>
						</table>
					</td>
				</tr>;
			default:
				return unrecognizedValue(secret.Kind);
			}
	}

	private fetchData() {
		const accountProm = getAccount(this.props.id);

		const folderProm = accountProm.then((acc) => getFolder(acc.FolderId));

		const secretsProm = accountProm.then((acc) => getSecrets(acc.Id));

		Promise.all([accountProm, folderProm, secretsProm]).then(([account, folderresponse, secrets]) => {
			this.setState({
				account,
				folderresponse,
				secrets,
			});
		}, defaultErrorHandler);
	}

	private getBreadcrumbItems(): Breadcrumb[] {
		const breadcrumbItems: Breadcrumb[] = [
			{ url: '', title: this.state.account.Title },
		];

		function unshiftFolderToBreadcrumb(fld: Folder) {
			breadcrumbItems.unshift({
				url: folderRoute.buildUrl({folderId: fld.Id}),
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
