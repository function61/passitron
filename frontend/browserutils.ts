// why these abstractions? it's not pretty putting these non-null asserts everywhere,
// and whether we should use document.location or window.location
// (or document.location.assign or document.location.href) is not something we want to
// repeat all around.

// returns "#foo" if hash present
// returns "" if no hash present
export function getCurrentHash(): string {
	return document.location!.hash;
}

// supports "#foo" (bare hash)
// supports "/path" (relative)
// supports "http://example.com/path" (absolute)
export function navigateTo(to: string): void {
	document.location!.assign(to);
}

export function reloadCurrentPage(): void {
	document.location!.reload();
}
