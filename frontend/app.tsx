import * as React from 'react';
import { router } from 'routes';

export interface AppProps {
	initialHash: string;
}

export interface AppState {
	hash: string;
}

export class App extends React.Component<AppProps, AppState> {
	private listenerProxy: any;

	constructor(props: AppProps) {
		super(props);

		this.state = {
			hash: props.initialHash,
		};
	}

	componentDidMount() {
		this.listenerProxy = () => {
			const newHash = document.location.hash;
			this.setState({ hash: newHash });
			return;
		};

		window.addEventListener('hashchange', this.listenerProxy);
	}

	componentWillUnmount() {
		window.removeEventListener('hashchange', this.listenerProxy);
	}

	render() {
		const fromRouter = router.match(window.location.hash);
		if (!fromRouter) {
			throw new Error('unknown page');
		}

		return fromRouter;
	}
}
