import { getCurrentHash } from 'f61ui/browserutils';
import { Breadcrumb } from 'f61ui/component/breadcrumbtrail';
import { NavLink } from 'f61ui/component/navigation';
import { DefaultLayout } from 'f61ui/layout/defaultlayout';
import { jsxChildType } from 'f61ui/types';
import { version } from 'generated/version';
import * as React from 'react';
import { auditlogRoute, indexRoute, settingsRoute, sshkeysRoute } from 'routes';

interface AppDefaultLayoutProps {
	title: string;
	breadcrumbs: Breadcrumb[];
	children: jsxChildType;
}

// app's default layout uses the default layout with props that are common to the whole app
export class AppDefaultLayout extends React.Component<AppDefaultLayoutProps, {}> {
	render() {
		const hash = getCurrentHash();

		const navLinks: NavLink[] = [
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

		return (
			<DefaultLayout
				appName="PiLockBox"
				appHomepage="https://github.com/function61/pi-security-module"
				navLinks={navLinks}
				logoUrl={indexRoute.buildUrl({})}
				breadcrumbs={this.props.breadcrumbs}
				content={this.props.children}
				version={version}
				pageTitle={this.props.title}
			/>
		);
	}
}
