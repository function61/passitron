
function make(...components: string[]): string {
	return '#' + components.join('/');
}

export function indexLink(): string {
	return make('index');
}

export function searchLink(query: string): string {
	return make('search', encodeURIComponent(query));
}

export function secretLink(id: string): string {
	return make('credview', id);
}

export function folderLink(id: string): string {
	return make('index', id);
}

export function sshKeysLink(): string {
	return make('sshkeys');
}

export function settingsLink(): string {
	return make('settings');
}

export function unsealLink(): string {
	return make('unseal');
}

export function auditLogLink(): string {
	return make('auditlog');
}

export function importOtpTokenLink(accountId: string): string {
	return make('importotptoken', accountId);
}
