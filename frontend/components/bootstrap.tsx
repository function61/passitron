import * as React from 'react';
import { jsxChildType } from 'types';

interface PanelProps {
	heading: string;
	children: jsxChildType;
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
