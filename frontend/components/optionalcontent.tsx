import * as React from 'react';

interface OptionalContentProps {
	children: string;
}

export class OptionalContent extends React.Component<OptionalContentProps, {}> {
	render() {
		return this.props.children ?
			<span>{this.props.children}</span> :
			<span className="text-muted">(Not set)</span>;
	}
}
