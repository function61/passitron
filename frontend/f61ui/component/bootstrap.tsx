import { jsxChildType } from 'f61ui/types';
import * as React from 'react';

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

interface ButtonProps {
	label: string;
	click: () => void;
}

export class Button extends React.Component<ButtonProps, {}> {
	render() {
		return (
			<span className="btn btn-default" onClick={this.props.click}>
				{this.props.label}
			</span>
		);
	}
}
