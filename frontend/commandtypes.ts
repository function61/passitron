
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
	Integer = 'integer',
}

export interface CommandField {
	Key: string;
	Required: boolean;
	Kind: CommandFieldKind;
	DefaultValueString?: string;
	DefaultValueBoolean?: boolean;
}
