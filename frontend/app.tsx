import { getCurrentHash } from 'browserutils';
import { DangerAlert } from 'components/alerts';
import { configureCsrfToken } from 'httputil';
import DefaultLayout from 'layouts/DefaultLayout';
import * as React from 'react';
import * as ReactDOM from 'react-dom';
import { router } from 'routes';

interface Config {
	csrf_token: string;
}

// entrypoint for the app. this is called when DOM is loaded
export function main(appElement: HTMLElement, config: Config): void {
	configureCsrfToken(config.csrf_token);

	ReactDOM.render(<App />, appElement);
}

export interface AppState {
	hash: string;
}

export class App extends React.Component<{}, AppState> {
	private hashChangedProxy: () => void;

	constructor(props: {}) {
		super(props);

		// need to create create bound proxy, because this object function
		// ref (bound one) must be used for removeEventListener()
		this.hashChangedProxy = () => {
			this.hashChanged();
		};

		this.state = {
			hash: getCurrentHash(),
		};
	}

	componentDidMount() {
		window.addEventListener('hashchange', this.hashChangedProxy);
	}

	componentWillUnmount() {
		window.removeEventListener('hashchange', this.hashChangedProxy);
	}

	render() {
		const fromRouter = router.match(this.state.hash);
		if (!fromRouter) {
			return (
				<DefaultLayout title="404" breadcrumbs={[]}>
					<h1>404</h1>

					<DangerAlert text="The page you were looking for is not found." />
				</DefaultLayout>
			);
		}

		return fromRouter;
	}

	private hashChanged() {
		this.setState({ hash: getCurrentHash() });
	}
}
