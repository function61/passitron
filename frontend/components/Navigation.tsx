import * as React from 'react';
import {indexRoute, sshkeysRoute, settingsRoute, auditlogRoute} from 'routes';

interface NavLink {
	url: string;
	title: string;
}

export default class Navigation extends React.Component<{}, {}> {
	render() {
		const links: NavLink[] = [
			{ title: 'Home', url: indexRoute.buildUrl({}) },
			{ title: 'SSH keys', url: sshkeysRoute.buildUrl({}) },
			{ title: 'Settings', url: settingsRoute.buildUrl({}) },
			{ title: 'Audit log', url: auditlogRoute.buildUrl({}) },
		];

		const items = links.map((link: NavLink) => {
			// TODO: <li class="active">
			return <li key={link.url}><a href={link.url}>{link.title}</a></li>;
		});

		return <ul className="nav nav-tabs">{items}</ul>;
	}
}
