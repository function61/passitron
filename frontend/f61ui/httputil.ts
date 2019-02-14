import { globalConfig } from 'f61ui/globalconfig';

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

	return fetch(url, {
		method: 'POST',
		headers: {
			Accept: 'application/json',
			'Content-Type': 'application/json',
			'x-csrf-token': globalConfig().csrfToken,
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
