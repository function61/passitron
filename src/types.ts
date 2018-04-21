
export interface CommandDefinition {
	title: string;
	key: string;
	fields: CommandField[];
}

export enum CommandFieldKind {
	Text = 'text',
	Password = 'password',
	Multiline = 'multiline',
	Checkbox = 'checkbox',
}

export interface CommandField {
	Key: string;
	Kind: CommandFieldKind;
	DefaultValueString?: string;
	DefaultValueBoolean?: boolean;
}
