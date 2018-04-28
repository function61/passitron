
export function unrecognizedValue(value: never): never {
	throw new Error(`Unrecognized value: ${value}`);
}

let uniqueDomIdCounter = 0;

export function uniqueDomId(): number {
	return ++uniqueDomIdCounter;
}
