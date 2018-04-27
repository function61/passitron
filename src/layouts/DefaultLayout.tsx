import * as React from 'react';
import Navigation from 'components/Navigation';
import {BreadcrumbTrail, Breadcrumb} from 'components/breadcrumbtrail';
import {indexLink} from 'links';

interface DefaultLayoutProps {
	title: string;
	breadcrumbs: Breadcrumb[];
	children: JSX.Element[] | JSX.Element;
}

export default class DefaultLayout extends React.Component<DefaultLayoutProps, {}> {
	render() {
		document.title = `${this.props.title} - PiLockBox`;

		return <div>
			<div className="header clearfix">
				<div className="pull-left">
					<h3 className="text-muted">
						<a href={indexLink()}>PiLockBox</a>
					</h3>
				</div>

				<nav className="pull-right">
					<Navigation />
				</nav>
			</div>

			<BreadcrumbTrail items={this.props.breadcrumbs} />

			{ this.props.children }
		</div>;
	}
}

