
export function getJson<T>(url: string): Promise<T> {
	const headers: {[key: string]: string} = {
		Accept: 'application/json',
	};

	const csrfToken = readCsrfToken();
	if (csrfToken) {
		headers['x-csrf-token'] = csrfToken;
	}

	return fetch(url, { method: 'GET', headers })
		.then(httpMustBeOk)
		.then((response) => response.json());
}

export function postJson<I, O>(url: string, body: I): Promise<O> {
	return postJsonReturningVoid<I>(url, body).then((res) => res.json());
}

export function postJsonReturningVoid<T>(url: string, body: T): Promise<Response> {
	const bodyToPost = JSON.stringify(body);

	const headers: {[key: string]: string} = {
		Accept: 'application/json',
		'Content-Type': 'application/json',
	};

	const csrfToken = readCsrfToken();
	if (csrfToken) {
		headers['x-csrf-token'] = csrfToken;
	}

	return fetch(url, {
		headers,
		method: 'POST',
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

function readCsrfToken(): string | null {
	// TODO: fix this botched way of reading the cookie value..
	const csrfToken = /csrf_token=([^;]+)/.exec(document.cookie);
	return csrfToken ? csrfToken[1] : null;
}
