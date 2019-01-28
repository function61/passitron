export interface U2FStdSignResult {
	keyHandle: string;
	clientData: string;
	signatureData: string;
}

export interface U2FStdRegisterRequest {
	version: string;
	challenge: string;
}

export interface U2FStdRegisteredKey {
	version: string;
	keyHandle: string;
	appId: string;
}

// TODO: how about error
export interface U2FStdRegisterResponse {
	registrationData: string;
	version: string;
	challenge: string;
	clientData: string;
}

export function u2fErrorMsg(resp: any): string {
	let msg = 'U2F error code ' + resp.errorCode;
	for (const name in u2f.ErrorCodes) {
		if (u2f.ErrorCodes[name] === resp.errorCode) {
			msg += ' (' + name + ')';
		}
	}

	if (resp.errorMessage) {
		msg += ': ' + resp.errorMessage;
	}

	return msg;
}

export function isU2FError(resp: any): boolean {
	if (!('errorCode' in resp)) {
		return false;
	}
	if (resp.errorCode === u2f.ErrorCodes.OK) {
		return false;
	}

	return true;
}
