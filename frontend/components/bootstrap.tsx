import * as React from 'react';

interface PanelProps {
	heading: string;
	children: JSX.Element[] | JSX.Element;
}

export class Panel extends React.Component<PanelProps, {}> {
	render() {
		return (
			<div className="panel panel-default">
				<div className="panel-heading">{this.props.heading}</div>
				<div className="panel-body">{this.props.children}</div>
			</div>
		);
	}
}
