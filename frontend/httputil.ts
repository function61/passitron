
export function getJson<T>(url: string): Promise<T> {
	return fetch(url)
		.then(httpMustBeOk)
		.then((response) => response.json());
}

export function httpMustBeOk(response: Response): Promise<Response> {
	if (!response.ok) {
		return new Promise<Response>((_: any, reject: any) => {
			response.text().then((text: string) => {
				if (response.headers.get('content-type') === 'application/json') {
					reject(JSON.parse(text) as {});
				} else {
					reject(new Error('HTTP response failure: ' + text));
				}
			}, (err: Error) => {
				reject(new Error('HTTP response failure. Also, error fetching response body: ' + err.toString()));
			});
		});
	}

	return Promise.resolve(response);
}
