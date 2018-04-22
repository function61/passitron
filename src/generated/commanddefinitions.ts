import {CommandDefinition, CommandFieldKind} from 'types';

// TODO: this will be a generated file

export function addOtpToken(accountId: string, otpToken: string): CommandDefinition {
	return {
		key: 'account.AddOtpToken',
		title: 'Add OTP token',
		fields: [
			{ Key: 'Account', Kind: CommandFieldKind.Text, DefaultValueString: accountId },
			{ Key: 'OtpProvisioningUrl', Kind: CommandFieldKind.Text, DefaultValueString: otpToken },
		],
	};
}

export function changeMasterPassword(): CommandDefinition {
	return {
		key: 'database.ChangeMasterPassword',
		title: 'Change master password',
		fields: [
			{ Key: 'NewMasterPassword', Kind: CommandFieldKind.Password },
			{ Key: 'NewMasterPasswordRepeat', Kind: CommandFieldKind.Password },
		],
	};
}

export function unseal(): CommandDefinition {
	return {
		key: 'database.Unseal',
		title: 'Unseal database',
		fields: [
			{ Key: 'MasterPassword', Kind: CommandFieldKind.Password },
		],
	};
}

export function writeKeepass(): CommandDefinition {
	return {
		key: 'database.ExportToKeepass',
		title: 'Export to KeePass format',
		fields: [ ],
	};
}

export function renameAccount(accountId: string, currentName: string): CommandDefinition {
	return {
		key: 'account.Rename',
		title: 'Rename account',
		fields: [
			{ Key: 'Account', Kind: CommandFieldKind.Text, DefaultValueString: accountId },
			{ Key: 'Title', Kind: CommandFieldKind.Text, DefaultValueString: currentName },
		],
	};
}

export function changeUsername(accountId: string, currentUsername: string): CommandDefinition {
	return {
		key: 'account.ChangeUsername',
		title: 'Change username',
		fields: [
			{ Key: 'Account', Kind: CommandFieldKind.Text, DefaultValueString: accountId },
			{ Key: 'Username', Kind: CommandFieldKind.Text, DefaultValueString: currentUsername },
		],
	};
}

export function changeDescription(accountId: string, currentDescription: string): CommandDefinition {
	return {
		key: 'account.ChangeDescription',
		title: 'Change description',
		fields: [
			{ Key: 'Account', Kind: CommandFieldKind.Text, DefaultValueString: accountId },
			{ Key: 'Description', Kind: CommandFieldKind.Multiline, DefaultValueString: currentDescription },
		],
	};
}

export function deleteSecret(accountId: string, secretId: string): CommandDefinition {
	return {
		key: 'account.DeleteSecret',
		title: 'Delete secret',
		fields: [
			{ Key: 'Account', Kind: CommandFieldKind.Text, DefaultValueString: accountId },
			{ Key: 'Secret', Kind: CommandFieldKind.Text, DefaultValueString: secretId },
		],
	};
}

export function createAccount(defaultFolderId: string): CommandDefinition {
	return {
		key: 'account.Create',
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
		key: 'account.CreateFolder',
		title: '+ Folder',
		fields: [
			{ Key: 'Parent', Kind: CommandFieldKind.Text, DefaultValueString: parentId },
			{ Key: 'Name', Kind: CommandFieldKind.Text },
		],
	};
}

export function renameFolder(folderId: string, currentName: string): CommandDefinition {
	return {
		key: 'account.RenameFolder',
		title: 'Rename',
		fields: [
			{ Key: 'Id', Kind: CommandFieldKind.Text, DefaultValueString: folderId },
			{ Key: 'Name', Kind: CommandFieldKind.Text, DefaultValueString: currentName },
		],
	};
}

export function moveFolder(folderId: string): CommandDefinition {
	return {
		key: 'account.MoveFolder',
		title: 'Move',
		fields: [
			{ Key: 'Id', Kind: CommandFieldKind.Text, DefaultValueString: folderId },
			{ Key: 'NewParent', Kind: CommandFieldKind.Text },
		],
	};
}

export function deleteAccount(accountId: string): CommandDefinition {
	return {
		key: 'account.Delete',
		title: 'Delete',
		fields: [
			{ Key: 'Id', Kind: CommandFieldKind.Text, DefaultValueString: accountId },
			{ Key: 'Confirm', Kind: CommandFieldKind.Checkbox, DefaultValueBoolean: false },
		],
	};
}

export function addPassword(accountId: string): CommandDefinition {
	return {
		key: 'account.AddPassword',
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
		key: 'account.AddSshKey',
		title: '+ SSH key',
		fields: [
			{ Key: 'Id', Kind: CommandFieldKind.Text, DefaultValueString: accountId },
			{ Key: 'SshPrivateKey', Kind: CommandFieldKind.Multiline },
		],
	};
}
