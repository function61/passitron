import clipboard from 'clipboard';
import {Breadcrumb} from 'components/breadcrumbtrail';
import {CommandIcon, CommandLink} from 'components/CommandButton';
import {Dropdown} from 'components/dropdown';
import {SecretReveal} from 'components/secretreveal';
import {Account, ExposedSecret, Folder, FolderResponse} from 'generated/apitypes';
import {
	AccountAddKeylist,
	AccountAddPassword,
	AccountAddSshKey,
	AccountChangeDescription,
	AccountChangeUsername,
	AccountDelete,
	AccountDeleteSecret,
	AccountRename,
} from 'generated/commanddefinitions';
import {SecretKind} from 'generated/domain';
import {defaultErrorHandler, getAccount, getFolder, getSecrets} from 'generated/restapi';
import DefaultLayout from 'layouts/DefaultLayout';
import * as React from 'react';
import {folderRoute, importotptokenRoute} from 'routes';
import {unrecognizedValue} from 'utils';

interface AccountPageProps {
	id: string;
}

interface AccountPageState {
	account: Account;
	secrets: ExposedSecret[];
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
				<Dropdown>
					<CommandLink command={AccountRename(account.Id, account.Title)} />
					<CommandLink command={AccountDelete(account.Id)} />

					<CommandLink command={AccountAddSshKey(account.Id)} />
					<CommandLink command={AccountAddKeylist(account.Id)} />
					<CommandLink command={AccountAddPassword(account.Id)} />

					<a href={importotptokenRoute.buildUrl({account: account.Id})}>+ OTP token</a>
				</Dropdown>
			</h1>

			<table className="table table-striped">
			<tbody>
				<tr>
					<td>
						Username
						<CommandIcon command={AccountChangeUsername(account.Id, account.Username)} />
					</td>
					<td>{account.Username}</td>
					<td onClick={() => this.copyToClipboard(account.Username)} className="fauxlink">📋</td>
				</tr>
				{secretRows}
				<tr>
					<td>
						Description
						<CommandIcon command={AccountChangeDescription(account.Id, account.Description)} />
					</td>
					<td style={{fontFamily: 'monospace', whiteSpace: 'pre'}}>{account.Description}</td>
					<td></td>
				</tr>
			</tbody>
			</table>
		</DefaultLayout>;
	}

	private secretToRow(exposedSecret: ExposedSecret, account: Account): JSX.Element {
		const secret = exposedSecret.Secret;

		switch (secret.Kind) {
			case SecretKind.SshKey:
				return <tr key={secret.Id}>
					<td>
						SSH public key
						<CommandIcon type="remove" command={AccountDeleteSecret(account.Id, secret.Id)} />
					</td>
					<td>{secret.SshPublicKeyAuthorized}</td>
					<td></td>
				</tr>;
			case SecretKind.Password:
				return <tr key={secret.Id}>
					<td>
						Password
						<CommandIcon type="remove" command={AccountDeleteSecret(account.Id, secret.Id)} />
					</td>
					<td><SecretReveal secret={secret.Password} /></td>
					<td onClick={() => this.copyToClipboard(secret.Password)} className="fauxlink">📋</td>
				</tr>;
			case SecretKind.OtpToken:
				return <tr key={secret.Id}>
					<td>
						OTP
						<CommandIcon type="remove" command={AccountDeleteSecret(account.Id, secret.Id)} />
					</td>
					<td>{exposedSecret.OtpProof}</td>
					<td onClick={() => this.copyToClipboard(exposedSecret.OtpProof)} className="fauxlink">📋</td>
				</tr>;
			case SecretKind.Keylist:
				const keyRows = secret.KeylistKeys.map((item) => <tr key={item.Key}>
					<th>{item.Key}</th>
					<td>{item.Value}</td>
				</tr>);

				return <tr key={secret.Id}>
					<td>
						Keylist
						<CommandIcon type="remove" command={AccountDeleteSecret(account.Id, secret.Id)} />
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
