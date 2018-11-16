import {defaultErrorHandler} from 'backenderrors';
import {elToClipboard} from 'clipboard';
import {DangerAlert} from 'components/alerts';
import {Breadcrumb} from 'components/breadcrumbtrail';
import {CommandIcon, CommandLink} from 'components/CommandButton';
import {Dropdown} from 'components/dropdown';
import {Loading} from 'components/loading';
import {MonospaceContent} from 'components/monospacecontent';
import {OptionalContent} from 'components/optionalcontent';
import {SecretReveal} from 'components/secretreveal';
import {
	Account,
	ExposedSecret,
	Folder,
	FolderResponse,
	Secret,
	SecretKeylistKey,
	U2FResponseBundle,
	U2FSignRequest,
	U2FSignResult,
	WrappedAccount,
} from 'generated/apitypes';
import {
	AccountAddKeylist,
	AccountAddPassword,
	AccountAddSecretNote,
	AccountAddSshKey,
	AccountChangeDescription,
	AccountChangeUrl,
	AccountChangeUsername,
	AccountDelete,
	AccountDeleteSecret,
	AccountRename,
} from 'generated/commanddefinitions';
import {SecretKind} from 'generated/domain';
import {getAccount, getFolder, getKeylistKey, getKeylistKeyChallenge, getSecrets} from 'generated/restapi';
import DefaultLayout from 'layouts/DefaultLayout';
import * as React from 'react';
import {folderRoute, importotptokenRoute} from 'routes';
import {isU2FError, u2fErrorMsg, U2FStdRegisteredKey, U2FStdSignResult} from 'u2ftypes';
import {relativeDateFormat, shouldAlwaysSucceed, unrecognizedValue} from 'utils';

interface SecretsFetcherProps {
	wrappedAccount: WrappedAccount;
	fetched: (secrets: ExposedSecret[]) => void;
}

interface SecretsFetcherState {
	authing: boolean;
	authError?: string;
}

class SecretsFetcher extends React.Component<SecretsFetcherProps, SecretsFetcherState> {
	state: SecretsFetcherState = { authing: false };

	componentDidMount() {
		// start fetching process automatically. in some rare cases the user might not
		// want this, but failed auth attempt timeouts are not dangerous and this reduces
		// extra clicks in the majority case
		shouldAlwaysSucceed(this.startSigning());
	}

	render() {
		if (this.state.authing) {
			return <div>
				<p>Please swipe your U2F token now ...</p>

				<Loading />
			</div>;
		}

		const authErrorNode = this.state.authError ?
			<DangerAlert text={this.state.authError} /> :
			'';

		return <div>
			<a className="btn btn-default" onClick={() => { shouldAlwaysSucceed(this.startSigning()); }}>
				Authenticate
			</a>

			{authErrorNode}
		</div>;
	}

	private async startSigning() {
		this.setState({ authing: true, authError: undefined });

		try {
			const result = await u2fSign(this.props.wrappedAccount.ChallengeBundle.SignRequest);

			if (isU2FError(result)) {
				this.setState({ authing: false, authError: u2fErrorMsg(result) });
				return;
			}

			const secrets = await getSecrets(this.props.wrappedAccount.Account.Id, {
				Challenge: this.props.wrappedAccount.ChallengeBundle.Challenge,
				SignResult: nativeSignResultToApiType(result),
			});

			this.props.fetched(secrets);
		} catch (e) {
			defaultErrorHandler(e);
		}
	}
}

// sign() errors are also resolved, but the value is an error value
async function u2fSign(req: U2FSignRequest): Promise<U2FStdSignResult> {
	return new Promise<U2FStdSignResult>((resolve) => {
		const keysTransformed: U2FStdRegisteredKey[] = req.RegisteredKeys.map((key) => {
			return {
				version: key.Version,
				keyHandle: key.KeyHandle,
				appId: key.AppID,
			};
		});

		u2f.sign(
			req.AppID,
			req.Challenge, // serialized (not in structural form)
			keysTransformed,
			(res: U2FStdSignResult) => { resolve(res); },
			5);
	});
}

interface KeylistAccessorProps {
	account: string;
	secret: Secret;
}

interface KeylistAccessorState {
	keylistKey: string;
	loading: boolean;
	authError?: string;
	foundKeyItem?: SecretKeylistKey;
}

class KeylistAccessor extends React.Component<KeylistAccessorProps, KeylistAccessorState> {
	state: KeylistAccessorState = { keylistKey: '', loading: false };

	render() {
		const authErrorNode = this.state.authError ?
			<DangerAlert text={this.state.authError} /> :
			'';

		const keyMaybe = this.state.foundKeyItem ?
			<div>
				<span className="label label-primary">{this.state.foundKeyItem.Value}</span>
				<span data-to-clipboard={this.state.foundKeyItem.Value} onClick={(e) => { elToClipboard(e); }} className="fauxlink margin-left">📋</span>
			</div> : null;

		return <div>
			<input className="form-control" style={{ width: '200px', display: 'inline-block' }} type="text" value={this.state.keylistKey} onChange={(e) => { this.onType(e); }} placeholder={this.props.secret.KeylistKeyExample} />

			<button className="btn btn-default" type="submit" onClick={() => { shouldAlwaysSucceed(this.onSubmit()); }}>Get</button>

			{authErrorNode}

			{this.state.loading ? <Loading /> : null}

			{keyMaybe}
		</div>;
	}

	private async onSubmit() {
		if (!this.state.keylistKey) {
			alert('no input');
			return;
		}

		// resetting foundKeyItem so if fetching multiple items, the old one does not
		// stay visible (which would confuse the user if it's the old or the new)
		this.setState({ loading: true, foundKeyItem: undefined, authError: undefined });

		try {
			const challengeBundle = await getKeylistKeyChallenge(this.props.account, this.props.secret.Id, this.state.keylistKey);

			const signResult = await u2fSign(challengeBundle.SignRequest);

			if (isU2FError(signResult)) {
				this.setState({ loading: false, authError: u2fErrorMsg(signResult) });
				return;
			}

			const challengeResponse: U2FResponseBundle = {
				SignResult: nativeSignResultToApiType(signResult),
				Challenge: challengeBundle.Challenge,
			};

			const foundKeyItem = await getKeylistKey(
				this.props.account,
				this.props.secret.Id,
				this.state.keylistKey,
				challengeResponse);

			this.setState({ foundKeyItem, loading: false });
		} catch (ex) {
			this.setState({ loading: false });
			defaultErrorHandler(ex);
		}
	}

	private onType(e: React.ChangeEvent<HTMLInputElement>) {
		this.setState({ keylistKey: e.target.value });
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
		shouldAlwaysSucceed(this.fetchData());
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
					<CommandLink command={AccountAddSecretNote(account.Id)} />

					<a href={importotptokenRoute.buildUrl({account: account.Id})}>+ OTP token</a>
				</Dropdown>
			</h1>

			<table className="table table-striped th-align-right">
			<tbody>
				<tr>
					<th>
						URL
						<CommandIcon command={AccountChangeUrl(account.Id, account.Url)} />
					</th>
					<td>{account.Url ? <a href={account.Url} target="_blank">{account.Url}</a> : <OptionalContent />}</td>
					<td></td>
				</tr>
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
					<td>
						<MonospaceContent><OptionalContent>{account.Description}</OptionalContent></MonospaceContent>
					</td>
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
				const exportUrl = `/accounts/${account.Id}/secrets/${secret.Id}/totp_barcode?mac=${exposedSecret.OtpKeyExportMac}`;

				return <tr key={secret.Id}>
					<th>
						<span title={relativeDateFormat(secret.Created)}>OTP code</span>
						<CommandIcon command={AccountDeleteSecret(account.Id, secret.Id)} />
					</th>
					<td>
						{exposedSecret.OtpProof}
						<a
							style={{marginLeft: '16px'}}
							title="Export to Google Authenticator"
							href={exportUrl}
							target="_blank"><span className="glyphicon glyphicon-barcode" /></a>
					</td>
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
			case SecretKind.Note:
				return <tr key={secret.Id}>
					<th>
						<span title={relativeDateFormat(secret.Created)}>Note</span>
						<CommandIcon command={AccountDeleteSecret(account.Id, secret.Id)} />
					</th>
					<td colSpan={2}>{secret.Title}
						<MonospaceContent>{secret.Note}</MonospaceContent>
					</td>
				</tr>;
			default:
				return unrecognizedValue(secret.Kind);
		}
	}

	private async fetchData() {
		const wrappedAccountProm = getAccount(this.props.id);

		const accountProm = wrappedAccountProm.then((wacc) => wacc.Account);

		const folderProm = accountProm.then((acc) => getFolder(acc.FolderId));

		const [wrappedAccount, account, folderresponse] = await Promise.all([wrappedAccountProm, accountProm, folderProm]);

		this.setState({
			wrappedAccount,
			account,
			folderresponse,
		});
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

function nativeSignResultToApiType(sr: U2FStdSignResult): U2FSignResult {
	return {
		KeyHandle: sr.keyHandle,
		SignatureData: sr.signatureData,
		ClientData: sr.clientData,
	};
}
