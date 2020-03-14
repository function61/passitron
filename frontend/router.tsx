import { makeRouter } from 'f61ui/typescript-safe-router/saferouter';
import { RootFolderId } from 'generated/domain_types';
import AccountPage from 'pages/AccountPage';
import AuditLogPage from 'pages/AuditLogPage';
import HomePage from 'pages/HomePage';
import ImportOtpToken from 'pages/ImportOtpToken';
import SearchPage from 'pages/SearchPage';
import SettingsPage from 'pages/SettingsPage';
import SignInPage from 'pages/SignInPage';
import SshKeysPage from 'pages/SshKeysPage';
import UnlockDecryptionKeyPage from 'pages/UnlockDecryptionKeyPage';
import * as React from 'react';
import * as r from 'routes';

export const router = makeRouter(r.indexRoute, () => (
	<HomePage key={RootFolderId} folderId={RootFolderId} />
))
	.registerRoute(r.folderRoute, (opts) => (
		<HomePage key={opts.folderId} folderId={opts.folderId} />
	))
	.registerRoute(r.searchRoute, (opts) => (
		<SearchPage key={opts.searchTerm} searchTerm={opts.searchTerm} />
	))
	.registerRoute(r.accountRoute, (opts) => <AccountPage key={opts.id} id={opts.id} />)
	.registerRoute(r.sshkeysRoute, () => <SshKeysPage />)
	.registerRoute(r.settingsRoute, () => <SettingsPage />)
	.registerRoute(r.signInRoute, (opts) => <SignInPage redirect={opts.redirect} />)
	.registerRoute(r.unlockDecryptionKeyRoute, (opts) => (
		<UnlockDecryptionKeyPage redirect={opts.redirect} />
	))
	.registerRoute(r.auditlogRoute, () => <AuditLogPage />)
	.registerRoute(r.importotptokenRoute, (opts) => (
		<ImportOtpToken key={opts.account} account={opts.account} />
	));
