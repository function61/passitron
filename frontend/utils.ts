import moment = require('moment');

export function unrecognizedValue(value: never): never {
	throw new Error(`Unrecognized value: ${value}`);
}

let uniqueDomIdCounter = 0;

export function uniqueDomId(): number {
	return ++uniqueDomIdCounter;
}

export function relativeDateFormat(dateIso: string): string {
	return moment(dateIso).fromNow();
}
