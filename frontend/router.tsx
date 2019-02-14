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
import * as React from 'react';
import {
	accountRoute,
	auditlogRoute,
	folderRoute,
	importotptokenRoute,
	indexRoute,
	searchRoute,
	settingsRoute,
	signInRoute,
	sshkeysRoute,
} from 'routes';

export const router = makeRouter(indexRoute, () => (
	<HomePage key={RootFolderId} folderId={RootFolderId} />
))
	.registerRoute(folderRoute, (opts) => <HomePage key={opts.folderId} folderId={opts.folderId} />)
	.registerRoute(searchRoute, (opts) => (
		<SearchPage key={opts.searchTerm} searchTerm={opts.searchTerm} />
	))
	.registerRoute(accountRoute, (opts) => <AccountPage key={opts.id} id={opts.id} />)
	.registerRoute(sshkeysRoute, () => <SshKeysPage />)
	.registerRoute(settingsRoute, () => <SettingsPage />)
	.registerRoute(signInRoute, (opts) => <SignInPage redirect={opts.redirect} />)
	.registerRoute(auditlogRoute, () => <AuditLogPage />)
	.registerRoute(importotptokenRoute, (opts) => (
		<ImportOtpToken key={opts.account} account={opts.account} />
	));
