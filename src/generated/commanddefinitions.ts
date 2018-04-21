import {CommandDefinition, CommandFieldKind} from 'types';

// TODO: this will be a generated file

export function changeMasterPassword(): CommandDefinition {
	return {
		key: 'ChangeMasterPasswordRequest',
		title: 'Change master password',
		fields: [
			{ Key: 'NewMasterPassword', Kind: CommandFieldKind.Password },
			{ Key: 'NewMasterPasswordRepeat', Kind: CommandFieldKind.Password },
		],
	};
}

export function createAccount(defaultFolderId: string): CommandDefinition {
	return {
		key: 'SecretCreateRequest',
		title: '+ Account',
		fields: [
			{ Key: 'FolderId', Kind: CommandFieldKind.Text, DefaultValueString: defaultFolderId },
			{ Key: 'Title', Kind: CommandFieldKind.Text },
			{ Key: 'Username', Kind: CommandFieldKind.Text },
			{ Key: 'Password', Kind: CommandFieldKind.Password },
		],
	};
}

export function createFolder(parentId: string): CommandDefinition {
	return {
		key: 'FolderCreateRequest',
		title: '+ Folder',
		fields: [
			{ Key: 'ParentId', Kind: CommandFieldKind.Text, DefaultValueString: parentId },
			{ Key: 'Name', Kind: CommandFieldKind.Text },
		],
	};
}

export function renameFolder(folderId: string, currentName: string): CommandDefinition {
	return {
		key: 'RenameFolderRequest',
		title: 'Rename',
		fields: [
			{ Key: 'Id', Kind: CommandFieldKind.Text, DefaultValueString: folderId },
			{ Key: 'Name', Kind: CommandFieldKind.Text, DefaultValueString: currentName },
		],
	};
}

export function moveFolder(folderId: string): CommandDefinition {
	return {
		key: 'MoveFolderRequest',
		title: 'Move',
		fields: [
			{ Key: 'Id', Kind: CommandFieldKind.Text, DefaultValueString: folderId },
			{ Key: 'NewParentId', Kind: CommandFieldKind.Text },
		],
	};
}

export function deleteAccount(accountId: string): CommandDefinition {
	return {
		key: 'DeleteAccountRequest',
		title: 'Delete',
		fields: [
			{ Key: 'Id', Kind: CommandFieldKind.Text, DefaultValueString: accountId },
			{ Key: 'Confirm', Kind: CommandFieldKind.Checkbox, DefaultValueBoolean: false },
		],
	};
}

export function addPassword(accountId: string): CommandDefinition {
	return {
		key: 'ChangePasswordRequest',
		title: '+ Password',
		fields: [
			{ Key: 'Id', Kind: CommandFieldKind.Text, DefaultValueString: accountId },
			{ Key: 'Password', Kind: CommandFieldKind.Password },
			{ Key: 'PasswordRepeat', Kind: CommandFieldKind.Password },
		],
	};
}

export function addSshKey(accountId: string): CommandDefinition {
	return {
		key: 'SetSshKeyRequest',
		title: '+ SSH key',
		fields: [
			{ Key: 'Account', Kind: CommandFieldKind.Text, DefaultValueString: accountId },
			{ Key: 'SshPrivateKey', Kind: CommandFieldKind.Multiline },
		],
	};
}
