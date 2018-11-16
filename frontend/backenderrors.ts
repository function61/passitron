import {unsealRoute} from 'routes';

export function defaultErrorHandler(err: Error | StructuredErrorResponse) {
	const ser = coerceToStructuredErrorResponse(err);

	if (handleDatabaseSealed(ser)) {
		return;
	}

	alert(`${ser.error_code}: ${ser.error_description}`);
}

export function isSealedError(err: StructuredErrorResponse): boolean {
	return err.error_code === 'database_is_sealed';
}

export function handleDatabaseSealed(err: StructuredErrorResponse): boolean {
	if (isSealedError(err)) {
		document.location.assign(unsealRoute.buildUrl({ redirect: document.location.hash }));
		return true;
	}

	return false;
}

export function coerceToStructuredErrorResponse(err: Error | StructuredErrorResponse): StructuredErrorResponse {
	if (isStructuredErrorResponse(err)) {
		return err;
	}

	return { error_code: 'generic_error', error_description: err.toString() };
}

export interface StructuredErrorResponse {
	error_code: string;
	error_description: string;
}

export function isStructuredErrorResponse(err: StructuredErrorResponse | {}): err is StructuredErrorResponse {
	return 'error_code' in (err as StructuredErrorResponse);
}
