import * as React from 'react';

export interface Breadcrumb {
	url: string;
	title: string;
}

interface BreadcrumbTrailProps {
	items: Breadcrumb[];
}

export class BreadcrumbTrail extends React.Component<BreadcrumbTrailProps, {}> {
	render() {
		const items = this.props.items.map((item, index) => {
			if (item.url === '') {
				return (
					<li key={index} className="active">
						{item.title}
					</li>
				);
			}
			return (
				<li key={index}>
					<a href={item.url}>{item.title}</a>
				</li>
			);
		});

		return <ol className="breadcrumb">{items}</ol>;
	}
}
