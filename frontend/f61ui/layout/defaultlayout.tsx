import { SearchBox } from 'components/SearchBox';
import NavigationTabs, { NavLink } from 'f61ui/component/navigation';
import { Breadcrumb, BreadcrumbTrail } from 'f61ui/components/breadcrumbtrail';
import { jsxChildType } from 'f61ui/types';
import * as React from 'react';

interface DefaultLayoutProps {
	pageTitle: string;
	version: string;
	appName: string;
	appHomepage: string;
	logoUrl: string;
	navLinks: NavLink[];
	breadcrumbs: Breadcrumb[];
	content: jsxChildType;
}

export class DefaultLayout extends React.Component<DefaultLayoutProps, {}> {
	render() {
		document.title = `${this.props.pageTitle} - ${this.props.appName}`;

		const dayOfWeek = [
			'Sunday',
			'Monday',
			'Tuesday',
			'Wednesday',
			'Thursday',
			'Friday',
			'Saturday',
		][new Date().getDay()];

		return (
			<div>
				<div className="header clearfix">
					<div className="pull-left">
						<h3 className="text-muted">
							<a href={this.props.logoUrl}>{this.props.appName}</a>
						</h3>
					</div>

					<div className="pull-left" style={{ padding: '14px 0 0 20px' }}>
						<SearchBox />
					</div>

					<nav className="pull-right">
						<NavigationTabs links={this.props.navLinks} />
					</nav>
				</div>

				<BreadcrumbTrail items={this.props.breadcrumbs} />

				{this.props.content}

				<div className="panel panel-default panel-footer" style={{ marginTop: '16px' }}>
					<div className="pull-left">
						<a href={this.props.appHomepage} target="_blank">
							{this.props.appName}
						</a>
						&nbsp;{this.props.version}&nbsp;by{' '}
						<a href="https://function61.com/" target="_blank">
							function61.com
						</a>
					</div>
					<div className="pull-right">Enjoy your {dayOfWeek}! :)</div>
					<div className="clearfix" />
				</div>
			</div>
		);
	}
}
