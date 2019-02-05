import { getCurrentHash } from 'f61ui/browserutils';
import * as React from 'react';
import { auditlogRoute, indexRoute, settingsRoute, sshkeysRoute } from 'routes';

interface NavLink {
	url: string;
	title: string;
	active: boolean;
}

export default class Navigation extends React.Component<{}, {}> {
	render() {
		const hash = getCurrentHash();

		const links: NavLink[] = [
			{
				title: 'Home',
				url: indexRoute.buildUrl({}),
				active: indexRoute.matchUrl(hash) !== null,
			},
			{
				title: 'SSH keys',
				url: sshkeysRoute.buildUrl({}),
				active: sshkeysRoute.matchUrl(hash) !== null,
			},
			{
				title: 'Settings',
				url: settingsRoute.buildUrl({}),
				active: settingsRoute.matchUrl(hash) !== null,
			},
			{
				title: 'Audit log',
				url: auditlogRoute.buildUrl({}),
				active: auditlogRoute.matchUrl(hash) !== null,
			},
		];

		const items = links.map((link) => {
			const activeOrNot = link.active ? 'active' : '';

			return (
				<li className={activeOrNot} key={link.url}>
					<a href={link.url}>{link.title}</a>
				</li>
			);
		});

		return <ul className="nav nav-tabs">{items}</ul>;
	}
}
