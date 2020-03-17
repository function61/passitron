import { U2FSignRequest, U2FSignResult } from 'generated/apitypes_types';

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

// sign() errors are also resolved (that is to say, this should never throw/reject), but
// the error is encapsulated inside U2FStdSignResult
export async function u2fSign(req: U2FSignRequest): Promise<U2FStdSignResult> {
	return new Promise<U2FStdSignResult>((resolve) => {
		const keysTransformed: U2FStdRegisteredKey[] = req.RegisteredKeys.map((key) => {
			return {
				version: key.Version,
				keyHandle: key.KeyHandle,
				appId: key.AppID,
			};
		});

		u2f.sign(
			req.AppID,
			req.Challenge, // serialized (not in structural form)
			keysTransformed,
			(res: U2FStdSignResult) => {
				resolve(res);
			},
			5,
		);
	});
}

export function nativeSignResultToApiType(sr: U2FStdSignResult): U2FSignResult {
	return {
		KeyHandle: sr.keyHandle,
		SignatureData: sr.signatureData,
		ClientData: sr.clientData,
	};
}
