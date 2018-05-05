import * as React from 'react';
import {makeRoute, makeRouter} from 'typescript-safe-router/saferouter';
import HomePage from 'pages/HomePage';
import SshKeysPage from 'pages/SshKeysPage';
import SearchPage from 'pages/SearchPage';
import ImportOtpToken from 'pages/ImportOtpToken';
import {RootFolderId} from 'generated/domain';
import AccountPage from 'pages/AccountPage';
import UnsealPage from 'pages/UnsealPage';
import AuditLogPage from 'pages/AuditLogPage';
import SettingsPage from 'pages/SettingsPage';

export const indexRoute = makeRoute('index', {});
export const folderRoute = makeRoute('folder', {folderId: 'string'});
export const searchRoute = makeRoute('search', {searchTerm: 'string'});
export const credviewRoute = makeRoute('credview', {id: 'string'});
export const sshkeysRoute = makeRoute('sshkeys', {});
export const settingsRoute = makeRoute('settings', {});
export const unsealRoute = makeRoute('unseal', {});
export const auditlogRoute = makeRoute('auditlog', {});
export const importotptokenRoute = makeRoute('importotptoken', {account: 'string'});

export const router = makeRouter(indexRoute, () => <HomePage key={RootFolderId} folderId={RootFolderId} />)
	.registerRoute(folderRoute, (opts) => <HomePage key={opts.folderId} folderId={opts.folderId} />)
	.registerRoute(searchRoute, (opts) => <SearchPage searchTerm={opts.searchTerm} />)
	.registerRoute(credviewRoute, (opts) => <AccountPage key={opts.id} id={opts.id} />)
	.registerRoute(sshkeysRoute, () => <SshKeysPage />)
	.registerRoute(settingsRoute, () => <SettingsPage />)
	.registerRoute(unsealRoute, () => <UnsealPage />)
	.registerRoute(auditlogRoute, () => <AuditLogPage />)
	.registerRoute(importotptokenRoute, (opts) => <ImportOtpToken account={opts.account} />);
