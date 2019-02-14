import { isNotSignedInError, isSealedError } from 'errors';
import { boot } from 'f61ui/appcontroller';
import { getCurrentHash, navigateTo } from 'f61ui/browserutils';
import { DangerAlert } from 'f61ui/component/alerts';
import { GlobalConfig } from 'f61ui/globalconfig';
import { StructuredErrorResponse } from 'f61ui/types';
import { AppDefaultLayout } from 'layout/appdefaultlayout';
import * as React from 'react';
import { router } from 'router';
import { signInRoute } from 'routes';

// entrypoint for the app. this is called when DOM is loaded
export function main(appElement: HTMLElement, config: GlobalConfig): void {
	config.knownGlobalErrorsHandler = (err: StructuredErrorResponse) => {
		if (isSealedError(err) || isNotSignedInError(err)) {
			navigateTo(signInRoute.buildUrl({ redirect: getCurrentHash() }));
			return true;
		}

		return false;
	};

	// AppController doesn't know how to use our custom app layout, so we have to tell it how
	// it would display a 404 page
	const notFoundPage = (
		<AppDefaultLayout title="404" breadcrumbs={[]}>
			<h1>404</h1>

			<DangerAlert text="The page you were looking for is not found." />
		</AppDefaultLayout>
	);

	boot(appElement, config, { router, notFoundPage });
}
