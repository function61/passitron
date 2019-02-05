import * as React from 'react';

export class MonospaceContent extends React.Component<{}, {}> {
	render() {
		return (
			<div style={{ fontFamily: 'monospace', whiteSpace: 'pre' }}>{this.props.children}</div>
		);
	}
}
