import { SearchBox } from 'components/SearchBox';
import { getCurrentLocation } from 'f61ui/browserutils';
import { Breadcrumb } from 'f61ui/component/breadcrumbtrail';
import { NavLink } from 'f61ui/component/navigation';
import { DefaultLayout } from 'f61ui/layout/defaultlayout';
import { version } from 'generated/version';
import * as React from 'react';
import { auditLogUrl, indexUrl, settingsUrl, sshKeysUrl } from 'generated/apitypes_uiroutes';

interface AppDefaultLayoutProps {
	title: string;
	breadcrumbs: Breadcrumb[];
	children: React.ReactNode;
}

// app's default layout uses the default layout with props that are common to the whole app
export class AppDefaultLayout extends React.Component<AppDefaultLayoutProps, {}> {
	render() {
		const currLoc = getCurrentLocation();

		const navLinks: NavLink[] = [
			{
				title: 'Home',
				url: indexUrl(),
				active: currLoc === indexUrl(),
			},
			{
				title: 'SSH keys',
				url: sshKeysUrl(),
				active: currLoc === sshKeysUrl(),
			},
			{
				title: 'Settings',
				url: settingsUrl(),
				active: currLoc === settingsUrl(),
			},
			{
				title: 'Audit log',
				url: auditLogUrl(),
				active: currLoc === auditLogUrl(),
			},
		];

		const appName = 'PiLockBox';
		return (
			<DefaultLayout
				appName={appName}
				appHomepage="https://github.com/function61/pi-security-module"
				navLinks={navLinks}
				logoNode={appName}
				logoClickUrl={indexUrl()}
				breadcrumbs={this.props.breadcrumbs}
				content={this.props.children}
				version={version}
				pageTitle={this.props.title}
				searchWidget={<SearchBox />}
			/>
		);
	}
}
