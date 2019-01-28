let csrfToken: undefined | string;

export function getJson<T>(url: string): Promise<T> {
	return fetch(url)
		.then(httpMustBeOk)
		.then((response) => response.json());
}

export function postJson<I, O>(url: string, body: I): Promise<O> {
	return postJsonReturningVoid<I>(url, body).then((res) => res.json());
}

export function postJsonReturningVoid<T>(url: string, body: T): Promise<Response> {
	const bodyToPost = JSON.stringify(body);

	if (!csrfToken) {
		return Promise.reject(new Error('csrfToken not set'));
	}

	return fetch(url, {
		method: 'POST',
		headers: {
			Accept: 'application/json',
			'Content-Type': 'application/json',
			'x-csrf-token': csrfToken,
		},
		body: bodyToPost,
	}).then(httpMustBeOk);
}

export function httpMustBeOk(response: Response): Promise<Response> {
	if (!response.ok) {
		return new Promise<Response>((_: any, reject: any) => {
			response.text().then(
				(text: string) => {
					if (response.headers.get('content-type') === 'application/json') {
						reject(JSON.parse(text) as {});
					} else {
						reject(new Error('HTTP response failure: ' + text));
					}
				},
				(err: Error) => {
					reject(
						new Error(
							'HTTP response failure. Also, error fetching response body: ' +
								err.toString(),
						),
					);
				},
			);
		});
	}

	return Promise.resolve(response);
}

export function configureCsrfToken(token: string) {
	if (csrfToken) {
		throw new Error('configureCsrfToken already called');
	}

	csrfToken = token;
}
