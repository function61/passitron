import * as React from 'react';

interface SecretRevealProps {
	secret: string;
}

interface SecretRevealState {
	visible: boolean;
}

export class SecretReveal extends React.Component<SecretRevealProps, SecretRevealState> {
	state = { visible: false };

	render() {
		const secretVisibleOrNot = this.state.visible ?
			this.props.secret :
			'********';

		const className = this.state.visible ?
			'glyphicon glyphicon-eye-close hovericon margin-left' :
			'glyphicon glyphicon-eye-open hovericon margin-left';

		return <span>
			{secretVisibleOrNot}
			<span
				className={className}
				onClick={() => { this.setState({visible: !this.state.visible}); }} />
		</span>;
	}
}
