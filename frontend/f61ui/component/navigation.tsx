import * as React from 'react';

export interface NavLink {
	url: string;
	title: string;
	active: boolean;
}

interface NavigationProps {
	links: NavLink[];
}

export default class NavigationTabs extends React.Component<NavigationProps, {}> {
	render() {
		const items = this.props.links.map((link) => {
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
