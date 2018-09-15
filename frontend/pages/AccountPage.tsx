import {elToClipboard} from 'clipboard';
import {Breadcrumb} from 'components/breadcrumbtrail';
import {CommandIcon, CommandLink} from 'components/CommandButton';
import {Dropdown} from 'components/dropdown';
import {Loading} from 'components/loading';
import {OptionalContent} from 'components/optionalcontent';
import {SecretReveal} from 'components/secretreveal';
import {
	Account,
	ExposedSecret,
	Folder,
	FolderResponse,
	Secret,
	SecretKeylistKey,
	WrappedAccount,
} from 'generated/apitypes';
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
import {defaultErrorHandler, getAccount, getFolder, getKeylistKey, getSecrets} from 'generated/restapi';
import DefaultLayout from 'layouts/DefaultLayout';
import * as React from 'react';
import {folderRoute, importotptokenRoute} from 'routes';
import {isU2FError, U2FStdRegisteredKey, U2FStdSignResult} from 'u2ftypes';
import {relativeDateFormat, unrecognizedValue} from 'utils';

interface SecretsFetcherProps {
	wrappedAccount: WrappedAccount;
	fetched: (secrets: ExposedSecret[]) => void;
}

interface SecretsFetcherState {
	authing: boolean;
}

class SecretsFetcher extends React.Component<SecretsFetcherProps, SecretsFetcherState> {
	state: SecretsFetcherState = { authing: false };

	render() {
		if (this.state.authing) {
			return <div>
				<p>Please swipe your U2F token now ...</p>

				<Loading />
			</div>;
		}

		return <a className="btn btn-default" onClick={() => { this.startSigning(); }}>Start</a>;
	}

	startSigning() {
		const sr = this.props.wrappedAccount.SignRequest;

		const signResult = (result: U2FStdSignResult) => {
			if (isU2FError(result)) {
				this.setState({ authing: false });
				return;
			}

			getSecrets(this.props.wrappedAccount.Account.Id, {
				Challenge: this.props.wrappedAccount.Challenge,
				SignResult: {
					KeyHandle: result.keyHandle,
					SignatureData: result.signatureData,
					ClientData: result.clientData,
				},
			}).then((secrets) => {
				this.props.fetched(secrets);
			}, defaultErrorHandler);
		};

		const keysTransformed: U2FStdRegisteredKey[] = sr.RegisteredKeys.map((key) => {
			return {
				version: key.Version,
				keyHandle: key.KeyHandle,
				appId: key.AppID,
			};
		});

		u2f.sign(
			sr.AppID,
			sr.Challenge, // serialized (not in structural form)
			keysTransformed,
			signResult,
			10);

		this.setState({ authing: true });
	}
}

interface KeylistAccessorProps {
	account: string;
	secret: Secret;
}

interface KeylistAccessorState {
	input: string;
	loading: boolean;
	foundKeyItem?: SecretKeylistKey;
}

class KeylistAccessor extends React.Component<KeylistAccessorProps, KeylistAccessorState> {
	state: KeylistAccessorState = { input: '', loading: false };

	render() {
		const keyMaybe = this.state.foundKeyItem ?
			<div>
				<span className="label label-primary">{this.state.foundKeyItem.Value}</span>
				<span data-to-clipboard={this.state.foundKeyItem.Value} onClick={(e) => { elToClipboard(e); }} className="fauxlink margin-left">📋</span>
			</div> : null;

		return <div>
			<input className="form-control" style={{ width: '200px', display: 'inline-block' }} type="text" value={this.state.input} onChange={(e) => { this.onType(e); }} placeholder={this.props.secret.KeylistKeyExample} />

			<button className="btn btn-default" type="submit" onClick={() => { this.onSubmit(); }}>Get</button>

			{this.state.loading ? <Loading /> : null}

			{keyMaybe}
		</div>;
	}

	private onSubmit() {
		if (!this.state.input) {
			alert('no input');
			return;
		}

		this.setState({ loading: true });

		getKeylistKey(this.props.account, this.props.secret.Id, this.state.input).then((foundKeyItem) => {
			this.setState({ foundKeyItem, loading: false });
		}, defaultErrorHandler);
	}

	private onType(e: React.ChangeEvent<HTMLInputElement>) {
		this.setState({ input: e.target.value });
	}
}

interface AccountPageProps {
	id: string;
}

interface AccountPageState {
	wrappedAccount: WrappedAccount;
	account: Account;
	secrets?: ExposedSecret[];
	folderresponse: FolderResponse;
}

export default class AccountPage extends React.Component<AccountPageProps, AccountPageState> {
	// https://developmentarc.gitbooks.io/react-indepth/content/life_cycle/the_life_cycle_recap.html
	componentDidMount() {
		this.fetchData();
	}

	render() {
		if (!this.state) {
			return <Loading />;
		}

		const account = this.state.account;

		const secretRows = this.state.secrets ?
			this.state.secrets.map((secret) => this.secretToRow(secret, account)) :
			<tr>
				<th>Secrets</th>
				<td>
					<SecretsFetcher
						wrappedAccount={this.state.wrappedAccount}
						fetched={(secrets) => { this.setState({ secrets }); }} />
				</td>
				<td></td>
			</tr>;

		const breadcrumbItems = this.getBreadcrumbItems();

		return <DefaultLayout title={account.Title} breadcrumbs={breadcrumbItems}>
			<h1>
				<span title={relativeDateFormat(account.Created)}>{account.Title}</span>
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

			<table className="table table-striped th-align-right">
			<tbody>
				<tr>
					<th>
						Username
						<CommandIcon command={AccountChangeUsername(account.Id, account.Username)} />
					</th>
					<td><OptionalContent>{account.Username}</OptionalContent></td>
					<td data-to-clipboard={account.Username} onClick={(e) => { elToClipboard(e); }} className="fauxlink">📋</td>
				</tr>
				{secretRows}
				<tr>
					<th>
						Description
						<CommandIcon command={AccountChangeDescription(account.Id, account.Description)} />
					</th>
					<td style={{fontFamily: 'monospace', whiteSpace: 'pre'}}><OptionalContent>{account.Description}</OptionalContent></td>
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
					<th>
						<span title={relativeDateFormat(secret.Created)}>SSH public key</span>
						<CommandIcon command={AccountDeleteSecret(account.Id, secret.Id)} />
					</th>
					<td>{secret.SshPublicKeyAuthorized}</td>
					<td></td>
				</tr>;
			case SecretKind.Password:
				return <tr key={secret.Id}>
					<th>
						<span title={relativeDateFormat(secret.Created)}>Password</span>
						<CommandIcon command={AccountDeleteSecret(account.Id, secret.Id)} />
					</th>
					<td><SecretReveal secret={secret.Password} /></td>
					<td data-to-clipboard={secret.Password} onClick={(e) => { elToClipboard(e); }} className="fauxlink">📋</td>
				</tr>;
			case SecretKind.OtpToken:
				return <tr key={secret.Id}>
					<th>
						<span title={relativeDateFormat(secret.Created)}>OTP code</span>
						<CommandIcon command={AccountDeleteSecret(account.Id, secret.Id)} />
					</th>
					<td>{exposedSecret.OtpProof}</td>
					<td data-to-clipboard={exposedSecret.OtpProof} onClick={(e) => { elToClipboard(e); }} className="fauxlink">📋</td>
				</tr>;
			case SecretKind.Keylist:
				return <tr key={secret.Id}>
					<th>
						<span title={relativeDateFormat(secret.Created)}>Keylist</span>
						<CommandIcon command={AccountDeleteSecret(account.Id, secret.Id)} />
					</th>
					<td colSpan={2}>{secret.Title}
						<KeylistAccessor account={account.Id} secret={secret} />
					</td>
				</tr>;
			default:
				return unrecognizedValue(secret.Kind);
		}
	}

	private fetchData() {
		const wrappedAccountProm = getAccount(this.props.id);

		const accountProm = wrappedAccountProm.then((wacc) => wacc.Account);

		const folderProm = accountProm.then((acc) => getFolder(acc.FolderId));

		Promise.all([wrappedAccountProm, accountProm, folderProm]).then(([wrappedAccount, account, folderresponse]) => {
			this.setState({
				wrappedAccount,
				account,
				folderresponse,
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
}
