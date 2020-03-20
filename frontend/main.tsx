import { isNotSignedInError, isSealedError } from 'errors';
import { boot, makeRouter } from 'f61ui/appcontroller';
import { getCurrentLocation, navigateTo } from 'f61ui/browserutils';
import { GlobalConfig } from 'f61ui/globalconfig';
import { StructuredErrorResponse } from 'f61ui/types';
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
import { DangerAlert } from 'f61ui/component/alerts';
import { AppDefaultLayout } from 'layout/appdefaultlayout';
import * as React from 'react';
import * as r from 'generated/apitypes_uiroutes';

class Handlers implements r.RouteHandlers {
	index() {
		return <HomePage key={RootFolderId} folderId={RootFolderId} />;
	}

	folder(opts: r.FolderOpts) {
		return <HomePage key={opts.id} folderId={opts.id} />;
	}

	search(opts: r.SearchOpts) {
		return <SearchPage key={opts.q} searchTerm={opts.q} />;
	}

	importOtpToken(opts: r.ImportOtpTokenOpts) {
		return <ImportOtpToken key={opts.account} account={opts.account} />;
	}

	account(opts: r.AccountOpts) {
		return <AccountPage key={opts.id} id={opts.id} />;
	}

	sshKeys() {
		return <SshKeysPage />;
	}

	settings() {
		return <SettingsPage />;
	}

	signIn(opts: r.SignInOpts) {
		return <SignInPage redirect={opts.next} />;
	}

	unlockDecryptionKey(opts: r.UnlockDecryptionKeyOpts) {
		return <UnlockDecryptionKeyPage redirect={opts.next} />;
	}

	auditLog() {
		return <AuditLogPage />;
	}

	notFound(url: string) {
		return (
			<AppDefaultLayout title="404" breadcrumbs={[]}>
				<h1>404</h1>

				<DangerAlert>The page ({url}) you were looking for, is not found.</DangerAlert>
			</AppDefaultLayout>
		);
	}
}

// entrypoint for the app. this is called when DOM is loaded
export function main(appElement: HTMLElement, config: GlobalConfig): void {
	config.knownGlobalErrorsHandler = (err: StructuredErrorResponse) => {
		if (isNotSignedInError(err)) {
			navigateTo(r.signInUrl({ next: getCurrentLocation() }));
			return true;
		} else if (isSealedError(err)) {
			navigateTo(r.unlockDecryptionKeyUrl({ next: getCurrentLocation() }));
			return true;
		}

		return false;
	};

	const handlers = new Handlers();

	boot(
		appElement,
		config,
		makeRouter(r.hasRouteFor, (relativeUrl: string) => r.handle(relativeUrl, handlers)),
	);
}
