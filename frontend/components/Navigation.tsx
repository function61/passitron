import * as React from 'react';
import {indexLink, sshKeysLink, settingsLink, auditLogLink} from 'links';

interface NavLink {
	url: string;
	title: string;
}

export default class Navigation extends React.Component<{}, {}> {
	render() {
		const links: NavLink[] = [
			{ title: 'Home', url: indexLink() },
			{ title: 'SSH keys', url: sshKeysLink() },
			{ title: 'Settings', url: settingsLink() },
			{ title: 'Audit log', url: auditLogLink() },
		];

		const items = links.map((link: NavLink) => {
			// TODO: <li class="active">
			return <li key={link.url}><a href={link.url}>{link.title}</a></li>;
		});

		return <ul className="nav nav-tabs">{items}</ul>;
	}
}
