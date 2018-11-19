import {RootFolderId} from 'generated/domain';
import AccountPage from 'pages/AccountPage';
import AuditLogPage from 'pages/AuditLogPage';
import HomePage from 'pages/HomePage';
import ImportOtpToken from 'pages/ImportOtpToken';
import SearchPage from 'pages/SearchPage';
import SettingsPage from 'pages/SettingsPage';
import SignInPage from 'pages/SignInPage';
import SshKeysPage from 'pages/SshKeysPage';
import * as React from 'react';
import {makeRoute, makeRouter} from 'typescript-safe-router/saferouter';

export const indexRoute = makeRoute('index', {});
export const folderRoute = makeRoute('folder', {folderId: 'string'});
export const searchRoute = makeRoute('search', {searchTerm: 'string'});
export const credviewRoute = makeRoute('credview', {id: 'string'});
export const sshkeysRoute = makeRoute('sshkeys', {});
export const settingsRoute = makeRoute('settings', {});
export const signInRoute = makeRoute('unseal', {redirect: 'string'});
export const auditlogRoute = makeRoute('auditlog', {});
export const importotptokenRoute = makeRoute('importotptoken', {account: 'string'});

export const router = makeRouter(indexRoute, () => <HomePage key={RootFolderId} folderId={RootFolderId} />)
	.registerRoute(folderRoute, (opts) => <HomePage key={opts.folderId} folderId={opts.folderId} />)
	.registerRoute(searchRoute, (opts) => <SearchPage key={opts.searchTerm} searchTerm={opts.searchTerm} />)
	.registerRoute(credviewRoute, (opts) => <AccountPage key={opts.id} id={opts.id} />)
	.registerRoute(sshkeysRoute, () => <SshKeysPage />)
	.registerRoute(settingsRoute, () => <SettingsPage />)
	.registerRoute(signInRoute, (opts) => <SignInPage redirect={opts.redirect} />)
	.registerRoute(auditlogRoute, () => <AuditLogPage />)
	.registerRoute(importotptokenRoute, (opts) => <ImportOtpToken key={opts.account} account={opts.account} />);
