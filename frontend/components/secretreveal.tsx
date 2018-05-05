import * as React from 'react';

interface SecretRevealProps {
	secret: string;
}

interface SecretRevealState {
	open: boolean;
}

export class SecretReveal extends React.Component<SecretRevealProps, SecretRevealState> {
	render() {
		if (this.state && this.state.open) {
			return <span>{this.props.secret}</span>;
		}

		return <span>
			********
			<span
				className="glyphicon glyphicon-eye-open hovericon margin-left"
				onClick={() => this.setState({open: true})} />
		</span>;
	}
}
