import * as React from 'react';
import * as ReactDOM from 'react-dom';
import { router } from 'routes';

// entrypoint for the app. this is called when DOM is loaded
export function main(appElement: HTMLElement): void {
	ReactDOM.render(
		<App />,
		appElement);
}

export interface AppState {
	hash: string;
}

export class App extends React.Component<{}, AppState> {
	private listenerProxy: any;

	constructor(props: {}) {
		super(props);

		// need to create create bound proxy, because this object function
		// ref (bound one) must be used for removeEventListener()
		this.listenerProxy = () => this.hashChanged();

		this.state = {
			hash: document.location.hash,
		};
	}

	componentDidMount() {
		window.addEventListener('hashchange', this.listenerProxy);
	}

	componentWillUnmount() {
		window.removeEventListener('hashchange', this.listenerProxy);
	}

	render() {
		const fromRouter = router.match(document.location.hash);
		if (!fromRouter) {
			throw new Error('unknown page');
		}

		return fromRouter;
	}

	private hashChanged() {
		this.setState({ hash: document.location.hash });
	}
}
