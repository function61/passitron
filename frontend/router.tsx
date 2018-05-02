import * as React from 'react';
import HomePage from 'pages/HomePage';
import SshKeysPage from 'pages/SshKeysPage';
import SearchPage from 'pages/SearchPage';
import ImportOtpToken from 'pages/ImportOtpToken';
import {rootFolderId} from 'model';
import AccountPage from 'pages/AccountPage';
import UnsealPage from 'pages/UnsealPage';
import AuditLogPage from 'pages/AuditLogPage';
import SettingsPage from 'pages/SettingsPage';

export interface RouterProps {
	initialHash: string;
}

export interface RouterState {
	hash: string;
}

export class Router extends React.Component<RouterProps, RouterState> {
	private listenerProxy: any;

	constructor(props: RouterProps) {
		super(props);
		this.state = {
			hash: props.initialHash,
		};
	}

	componentDidMount() {
		this.listenerProxy = () => {
			var newHash = document.location.hash;
			this.setState({ hash: newHash });
			return;
		};

		window.addEventListener('hashchange', this.listenerProxy);
	}

	componentWillUnmount() {
		window.removeEventListener('hashchange', this.listenerProxy);
	}

	render() {
		let page: JSX.Element | null = null;

		const hash = this.state.hash === '' ?
			[] :
			this.state.hash.substr(1).split('/');


		if (hash.length === 0) {
			throw new Error('unknown page');
		}

		var firstComponent = hash[0];

		if (firstComponent === 'index' && hash.length === 1) {
			page = <HomePage key={rootFolderId} folderId={rootFolderId} />;
		} else if (firstComponent === 'index' && hash.length === 2) {
			page = <HomePage key={rootFolderId} folderId={hash[1]} />;
		} else if (firstComponent === 'search' && hash.length === 2) {
			const searchTerm = decodeURIComponent(hash[1]);

			page = <SearchPage searchTerm={searchTerm} />;
		} else if (firstComponent === 'credview' && hash.length === 2) {
			page = <AccountPage id={hash[1]} />;
		} else if (firstComponent === 'sshkeys' && hash.length === 1) {
			page = <SshKeysPage />;
		} else if (firstComponent === 'settings' && hash.length === 1) {
			page = <SettingsPage />;
		} else if (firstComponent === 'unseal' && hash.length === 1) {
			page = <UnsealPage />;
		} else if (firstComponent === 'auditlog' && hash.length === 1) {
			page = <AuditLogPage />;
		} else if (firstComponent === 'importotptoken' && hash.length === 2) {
			page = <ImportOtpToken account={hash[1]} />;
		} else {
			throw new Error('unknown page');
		}

		return page;
	}
}
