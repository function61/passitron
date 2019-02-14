import { makeRoute } from 'f61ui/typescript-safe-router/saferouter';

export const indexRoute = makeRoute('index', {});
export const folderRoute = makeRoute('folder', { folderId: 'string' });
export const searchRoute = makeRoute('search', { searchTerm: 'string' });
export const accountRoute = makeRoute('account', { id: 'string' });
export const sshkeysRoute = makeRoute('sshkeys', {});
export const settingsRoute = makeRoute('settings', {});
export const signInRoute = makeRoute('unseal', { redirect: 'string' });
export const auditlogRoute = makeRoute('auditlog', {});
export const importotptokenRoute = makeRoute('importotptoken', { account: 'string' });
