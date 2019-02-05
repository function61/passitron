import * as React from 'react';

export class OptionalContent extends React.Component<{}, {}> {
	render() {
		return this.props.children ? (
			<span>{this.props.children}</span>
		) : (
			<span className="text-muted">(Not set)</span>
		);
	}
}
