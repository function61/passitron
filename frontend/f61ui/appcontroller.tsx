import { getCurrentHash } from 'f61ui/browserutils';
import { configureCsrfToken } from 'f61ui/httputil';
import { Router } from 'f61ui/typescript-safe-router/saferouter';
import * as React from 'react';
import * as ReactDOM from 'react-dom';

export interface AppControllerConfig {
	csrf_token: string;
}

// entrypoint for the app. this is called when DOM is loaded
export function boot(
	appElement: HTMLElement,
	config: AppControllerConfig,
	props: AppControllerProps,
): void {
	configureCsrfToken(config.csrf_token);

	ReactDOM.render(<AppController {...props} />, appElement);
}

interface AppControllerProps {
	router: Router<JSX.Element>;
	notFoundPage: JSX.Element;
}

export interface AppControllerState {
	hash: string;
}

export class AppController extends React.Component<AppControllerProps, AppControllerState> {
	state: AppControllerState = { hash: getCurrentHash() };

	componentDidMount() {
		window.addEventListener('hashchange', this.hashChangedProxy);
	}

	componentWillUnmount() {
		window.removeEventListener('hashchange', this.hashChangedProxy);
	}

	render() {
		const fromRouter = this.props.router.match(this.state.hash);
		if (!fromRouter) {
			return this.props.notFoundPage;
		}

		return fromRouter;
	}

	// need to create create bound proxy, because this object function
	// ref (bound one) must be used for removeEventListener()
	private hashChangedProxy = () => {
		this.hashChanged();
	};

	private hashChanged() {
		this.setState({ hash: getCurrentHash() });
	}
}
