import 'bootstrap'; // side effect import, cool stuff guys
import { getCurrentHash } from 'f61ui/browserutils';
import { GlobalConfig, globalConfigure } from 'f61ui/globalconfig';
import { Router } from 'f61ui/typescript-safe-router/saferouter';
import * as React from 'react';
import * as ReactDOM from 'react-dom';

// entrypoint for the app. this is called when DOM is loaded
export function boot(
	appElement: HTMLElement,
	globalConfig: GlobalConfig,
	props: AppControllerProps,
): void {
	globalConfigure(globalConfig);

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
