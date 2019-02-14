import { StructuredErrorResponse } from 'f61ui/types';

export function isSealedError(err: StructuredErrorResponse): boolean {
	return err.error_code === 'database_is_sealed';
}

export function isNotSignedInError(err: StructuredErrorResponse): boolean {
	return err.error_code === 'not_signed_in';
}
