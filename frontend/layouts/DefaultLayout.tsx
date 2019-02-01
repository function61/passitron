import { Breadcrumb, BreadcrumbTrail } from 'components/breadcrumbtrail';
import Navigation from 'components/Navigation';
import { SearchBox } from 'components/SearchBox';
import { version } from 'generated/version';
import * as React from 'react';
import { indexRoute } from 'routes';
import { jsxChildType } from 'types';

interface DefaultLayoutProps {
	title: string;
	breadcrumbs: Breadcrumb[];
	children: jsxChildType;
}

export default class DefaultLayout extends React.Component<DefaultLayoutProps, {}> {
	render() {
		document.title = `${this.props.title} - PiLockBox`;

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
							<a href={indexRoute.buildUrl({})}>PiLockBox</a>
						</h3>
					</div>

					<div className="pull-left" style={{ padding: '14px 0 0 20px' }}>
						<SearchBox />
					</div>

					<nav className="pull-right">
						<Navigation />
					</nav>
				</div>

				<BreadcrumbTrail items={this.props.breadcrumbs} />

				{this.props.children}

				<div className="panel panel-default panel-footer" style={{ marginTop: '16px' }}>
					<div className="pull-left">
						<a href="https://github.com/function61/pi-security-module" target="_blank">
							PiLockBox
						</a>
						&nbsp;{version}&nbsp;by{' '}
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
