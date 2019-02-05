import { Breadcrumb } from 'f61ui/components/breadcrumbtrail';
import { DefaultLayout } from 'f61ui/layout/defaultlayout';
import { jsxChildType } from 'f61ui/types';
import { version } from 'generated/version';
import * as React from 'react';
import { indexRoute } from 'routes';

interface AppDefaultLayoutProps {
	title: string;
	breadcrumbs: Breadcrumb[];
	children: jsxChildType;
}

export class AppDefaultLayout extends React.Component<AppDefaultLayoutProps, {}> {
	render() {
		return (
			<DefaultLayout
				appName="PiLockBox"
				appHomepage="https://github.com/function61/pi-security-module"
				logoUrl={indexRoute.buildUrl({})}
				breadcrumbs={this.props.breadcrumbs}
				content={this.props.children}
				version={version}
				pageTitle={this.props.title}
			/>
		);
	}
}
