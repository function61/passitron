import { getCurrentHash } from 'f61ui/browserutils';
import { DangerAlert } from 'f61ui/components/alerts';
import { configureCsrfToken } from 'f61ui/httputil';
import { AppDefaultLayout } from 'layout/appdefaultlayout';
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
				<AppDefaultLayout title="404" breadcrumbs={[]}>
					<h1>404</h1>

					<DangerAlert text="The page you were looking for is not found." />
				</AppDefaultLayout>
			);
		}

		return fromRouter;
	}

	private hashChanged() {
		this.setState({ hash: getCurrentHash() });
	}
}
