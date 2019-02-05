export enum CrudNature {
	update = 'update',
	delete = 'delete',
	create = 'create',
}

export interface CommandDefinition {
	title: string;
	key: string;
	additional_confirmation?: string;
	crudNature: CrudNature;
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
	HideIfDefaultValue: boolean;
	Kind: CommandFieldKind;
	DefaultValueString?: string;
	DefaultValueBoolean?: boolean;
	Help?: string;
	ValidationRegex: string;
}
