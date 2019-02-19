import { StructuredErrorResponse } from 'f61ui/types';

export interface GlobalConfig {
	assetsDir: string;
	knownGlobalErrorsHandler?: (err: StructuredErrorResponse) => boolean;
}

let gConv: GlobalConfig | undefined;

export function globalConfigure(conf: GlobalConfig) {
	gConv = conf;
}

export function globalConfig(): GlobalConfig {
	if (!gConv) {
		throw new Error('not configured');
	}

	return gConv;
}
