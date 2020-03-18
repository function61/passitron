import { U2fSigner } from 'components/U2F';
import { PrimaryLabel } from 'f61ui/component/bootstrap';
import { Breadcrumb } from 'f61ui/component/breadcrumbtrail';
import { ClipboardButton } from 'f61ui/component/clipboardbutton';
import { CommandIcon, CommandLink } from 'f61ui/component/CommandButton';
import { Dropdown } from 'f61ui/component/dropdown';
import { Loading } from 'f61ui/component/loading';
import { MonospaceContent } from 'f61ui/component/monospacecontent';
import { OptionalContent } from 'f61ui/component/optionalcontent';
import { Result } from 'f61ui/component/result';
import { SecretReveal } from 'f61ui/component/secretreveal';
import { relativeDateFormat, shouldAlwaysSucceed, unrecognizedValue } from 'f61ui/utils';
import {
	AccountAddExternalU2FToken,
	AccountAddExternalYubicoOtpToken,
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
} from 'generated/apitypes_commands';
import {
	getAccount,
	getFolder,
	getKeylistItem,
	getKeylistItemChallenge,
	getSecrets,
	totpBarcodeExportUrl,
} from 'generated/apitypes_endpoints';
import {
	Account,
	ExposedSecret,
	Folder,
	FolderResponse,
	Secret,
	SecretKeylistKey,
	U2FChallengeBundle,
	WrappedAccount,
} from 'generated/apitypes_types';
import { ExternalTokenKind, SecretKind } from 'generated/domain_types';
import { AppDefaultLayout } from 'layout/appdefaultlayout';
import * as React from 'react';
import { folderRoute, importotptokenRoute } from 'routes';

interface SecretsFetcherProps {
	wrappedAccount: WrappedAccount;
	fetched: (secrets: ExposedSecret[]) => void;
}

interface SecretsFetcherState {
	secrets?: Result<ExposedSecret[]>;
}

class SecretsFetcher extends React.Component<SecretsFetcherProps, SecretsFetcherState> {
	state: SecretsFetcherState = {};

	render() {
		if (this.state.secrets) {
			const [, loadingOrError] = this.state.secrets.unwrap();
			return loadingOrError;
		}

		return (
			<U2fSigner
				challenge={this.props.wrappedAccount.ChallengeBundle}
				signed={(signature) => {
					const secrets = new Result<ExposedSecret[]>((x) => {
						this.setState({ secrets: x });
					});

					this.setState({ secrets });

					const secretsProm = getSecrets(this.props.wrappedAccount.Account.Id, signature);

					secrets.load(() => secretsProm);

					secretsProm.then(
						(secretsRaw) => {
							this.props.fetched(secretsRaw);
						},
						() => {
							// errors already handled in another handler
						},
					);
				}}
			/>
		);
	}
}

interface KeylistAccessorProps {
	account: string;
	secret: Secret;
}

interface KeylistAccessorState {
	keylistKey: string;
	challenge: Result<U2FChallengeBundle>;
	foundKeyItem: Result<SecretKeylistKey>;
}

class KeylistAccessor extends React.Component<KeylistAccessorProps, KeylistAccessorState> {
	state: KeylistAccessorState = {
		keylistKey: '',
		challenge: new Result<U2FChallengeBundle>((x) => {
			this.setState({ challenge: x });
		}),
		foundKeyItem: new Result<SecretKeylistKey>((x) => {
			this.setState({ foundKeyItem: x });
		}),
	};

	render() {
		return (
			<div>
				<input
					className="form-control"
					style={{ width: '200px', display: 'inline-block' }}
					type="text"
					value={this.state.keylistKey}
					onChange={(e) => {
						this.onType(e);
					}}
					placeholder={this.props.secret.KeylistKeyExample}
				/>

				<button
					className="btn btn-default"
					type="submit"
					onClick={() => {
						shouldAlwaysSucceed(this.onSubmit());
					}}>
					Get
				</button>

				{this.state.foundKeyItem.draw((foundKeyItem) => (
					<div>
						<PrimaryLabel>{foundKeyItem.Value}</PrimaryLabel>
						<ClipboardButton text={foundKeyItem.Value} />
					</div>
				))}

				{this.state.challenge.draw((challenge) => (
					<U2fSigner
						challenge={challenge}
						signed={(signature) => {
							// signer has done its job
							this.state.challenge.reset();

							this.state.foundKeyItem.load(() =>
								getKeylistItem(
									this.props.account,
									this.props.secret.Id,
									this.state.keylistKey,
									signature,
								),
							);
						}}
					/>
				))}
			</div>
		);
	}

	private async onSubmit() {
		// resetting these so if fetching multiple items, the old one does not stay
		// visible to confuse the user
		this.state.challenge.reset();
		this.state.foundKeyItem.reset();

		if (!this.state.keylistKey) {
			alert('no input');
			return;
		}

		this.state.challenge.load(() =>
			getKeylistItemChallenge(
				this.props.account,
				this.props.secret.Id,
				this.state.keylistKey,
			),
		);
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

		const secretRows = this.state.secrets ? (
			this.state.secrets.map((secret) => this.secretToRow(secret, account))
		) : (
			<tr>
				<th>Secrets</th>
				<td>
					<SecretsFetcher
						wrappedAccount={this.state.wrappedAccount}
						fetched={(secrets) => {
							this.setState({ secrets });
						}}
					/>
				</td>
				<td />
			</tr>
		);

		const breadcrumbItems = this.getBreadcrumbItems();

		return (
			<AppDefaultLayout title={account.Title} breadcrumbs={breadcrumbItems}>
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
						<CommandLink command={AccountAddExternalU2FToken(account.Id)} />
						<CommandLink command={AccountAddExternalYubicoOtpToken(account.Id)} />

						<a href={importotptokenRoute.buildUrl({ account: account.Id })}>
							+ OTP token
						</a>
					</Dropdown>
				</h1>

				<table className="table table-striped th-align-right">
					<tbody>
						<tr>
							<th>
								URL
								<CommandIcon command={AccountChangeUrl(account.Id, account.Url)} />
							</th>
							<td>
								{account.Url ? (
									<a href={account.Url} target="_blank">
										{account.Url}
									</a>
								) : (
									<OptionalContent />
								)}
							</td>
							<td />
						</tr>
						<tr>
							<th>
								Username
								<CommandIcon
									command={AccountChangeUsername(account.Id, account.Username)}
								/>
							</th>
							<td>
								<OptionalContent>{account.Username}</OptionalContent>
							</td>
							<td>
								<ClipboardButton text={account.Username} />
							</td>
						</tr>
						{secretRows}
						<tr>
							<th>
								Description
								<CommandIcon
									command={AccountChangeDescription(
										account.Id,
										account.Description,
									)}
								/>
							</th>
							<td>
								<MonospaceContent>
									<OptionalContent>{account.Description}</OptionalContent>
								</MonospaceContent>
							</td>
							<td />
						</tr>
					</tbody>
				</table>
			</AppDefaultLayout>
		);
	}

	private secretToRow(exposedSecret: ExposedSecret, account: Account): JSX.Element {
		const secret = exposedSecret.Secret;

		switch (secret.Kind) {
			case SecretKind.SshKey:
				return (
					<tr key={secret.Id}>
						<th>
							<span title={relativeDateFormat(secret.Created)}>SSH public key</span>
							<CommandIcon command={AccountDeleteSecret(account.Id, secret.Id)} />
						</th>
						<td>{secret.SshPublicKeyAuthorized}</td>
						<td />
					</tr>
				);
			case SecretKind.Password:
				return (
					<tr key={secret.Id}>
						<th>
							<span title={relativeDateFormat(secret.Created)}>Password</span>
							<CommandIcon command={AccountDeleteSecret(account.Id, secret.Id)} />
						</th>
						<td>
							<SecretReveal secret={secret.Password} />
						</td>
						<td>
							<ClipboardButton text={secret.Password} />
						</td>
					</tr>
				);
			case SecretKind.OtpToken:
				const exportUrl = totpBarcodeExportUrl(
					account.Id,
					secret.Id,
					exposedSecret.OtpKeyExportMac,
				);

				return (
					<tr key={secret.Id}>
						<th>
							<span title={relativeDateFormat(secret.Created)}>OTP code</span>
							<CommandIcon command={AccountDeleteSecret(account.Id, secret.Id)} />
						</th>
						<td>
							{exposedSecret.OtpProof}
							<a
								style={{ marginLeft: '16px' }}
								title="Export to Google Authenticator"
								href={exportUrl}
								target="_blank">
								<span className="glyphicon glyphicon-barcode" />
							</a>
						</td>
						<td>
							<ClipboardButton text={exposedSecret.OtpProof} />
						</td>
					</tr>
				);
			case SecretKind.Keylist:
				return (
					<tr key={secret.Id}>
						<th>
							<span title={relativeDateFormat(secret.Created)}>Keylist</span>
							<CommandIcon command={AccountDeleteSecret(account.Id, secret.Id)} />
						</th>
						<td colSpan={2}>
							{secret.Title}
							<KeylistAccessor account={account.Id} secret={secret} />
						</td>
					</tr>
				);
			case SecretKind.ExternalToken:
				return (
					<tr key={secret.Id}>
						<th>
							<span title={relativeDateFormat(secret.Created)}>
								{externalTokenKindHumanReadable(secret.ExternalTokenKind!)}{' '}
								(external token)
							</span>
							<CommandIcon command={AccountDeleteSecret(account.Id, secret.Id)} />
						</th>
						<td colSpan={2}>{secret.Title}</td>
					</tr>
				);
			case SecretKind.Note:
				return (
					<tr key={secret.Id}>
						<th>
							<span title={relativeDateFormat(secret.Created)}>Note</span>
							<CommandIcon command={AccountDeleteSecret(account.Id, secret.Id)} />
						</th>
						<td colSpan={2}>
							{secret.Title}
							<MonospaceContent>{secret.Note}</MonospaceContent>
						</td>
					</tr>
				);
			default:
				return unrecognizedValue(secret.Kind);
		}
	}

	private async fetchData() {
		const wrappedAccountProm = getAccount(this.props.id);

		const accountProm = wrappedAccountProm.then((wacc) => wacc.Account);

		const folderProm = accountProm.then((acc) => getFolder(acc.FolderId));

		const [wrappedAccount, account, folderresponse] = await Promise.all([
			wrappedAccountProm,
			accountProm,
			folderProm,
		]);

		this.setState({
			wrappedAccount,
			account,
			folderresponse,
		});
	}

	private getBreadcrumbItems(): Breadcrumb[] {
		const breadcrumbItems: Breadcrumb[] = [{ url: '', title: this.state.account.Title }];

		function unshiftFolderToBreadcrumb(fld: Folder) {
			breadcrumbItems.unshift({
				url: folderRoute.buildUrl({ folderId: fld.Id }),
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

function externalTokenKindHumanReadable(kind: ExternalTokenKind): string {
	switch (kind) {
		case ExternalTokenKind.U2f:
			return 'U2F';
		case ExternalTokenKind.YubicoOtp:
			return 'Yubico OTP';
		default:
			return unrecognizedValue(kind);
	}
}
