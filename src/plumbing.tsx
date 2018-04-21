import * as React from 'react';
import {CommandDefinition, CommandField, CommandFieldKind} from 'types';
import {httpMustBeOk} from 'repo';
import {unrecognizedValue} from 'utils';

export type CommandFieldChangeListener = (key: string, value: string | boolean) => void;

export class CommandManager {
	private definition: CommandDefinition;
	private commandBody: {[key: string]: any} = {};

	constructor(definition: CommandDefinition) {
		this.definition = definition;

		this.definition.fields.forEach((field) => {
			switch (field.Kind) {
			case CommandFieldKind.Password:
			case CommandFieldKind.Text:
			case CommandFieldKind.Multiline:
				this.commandBody[field.Key] = field.DefaultValueString;
				break;
			case CommandFieldKind.Checkbox:
				this.commandBody[field.Key] = field.DefaultValueBoolean;
				break;
			default:
				unrecognizedValue(field.Kind);
			}
		});
	}

	getChangeHandlerBound(): CommandFieldChangeListener {
		return this.changeHandler.bind(this);
	}

	execute(): Promise<void> {
		const bodyToPost = JSON.stringify(this.commandBody);

		return fetch(`/command/${this.definition.key}`, {
			method: 'POST',
			headers: {
				Accept: 'application/json',
				'Content-Type': 'application/json',
			},
			body: bodyToPost,
		})
			.then(httpMustBeOk)
			.then(res => res.json())
			.then(_ => {
				return;
			});
	}

	private changeHandler(key: string, value: string | boolean) {
		this.commandBody[key] = value;
	}
}

interface CommandPageletProps {
	command: CommandDefinition;
	updates: CommandFieldChangeListener;
}

export class CommandPagelet extends React.Component<CommandPageletProps, {}> {
	render() {
		const fieldGroups = this.props.command.fields.map((field, idx) => {
			const input = this.createInput(field);

			return <div className="form-group" key={idx}>
				<label>{field.Key}</label>

				{input}
			</div>;
		});

		return <form>
			{fieldGroups}
		</form>;
	}

	private onInputChange(e: React.FormEvent<HTMLInputElement>) {
		this.props.updates(e.currentTarget.name, e.currentTarget.value);
	}

	private onCheckboxChange(e: React.FormEvent<HTMLInputElement>) {
		this.props.updates(e.currentTarget.name, e.currentTarget.checked);
	}

	private onTextareaChange(e: React.FormEvent<HTMLTextAreaElement>) {
		this.props.updates(e.currentTarget.name, e.currentTarget.value);
	}

	private createInput(field: CommandField): JSX.Element {
		switch (field.Kind) {
		case CommandFieldKind.Password:
			return <input
				type="password"
				className="form-control"
				name={field.Key}
				onChange={this.onInputChange.bind(this)}
			/>;
		case CommandFieldKind.Text:
			return <input
				type="text"
				className="form-control"
				name={field.Key}
				defaultValue={field.DefaultValueString}
				onChange={this.onInputChange.bind(this)}
			/>;
		case CommandFieldKind.Multiline:
			return <textarea
				name={field.Key}
				className="form-control"
				rows={7}
				defaultValue={field.DefaultValueString}
				onChange={this.onTextareaChange.bind(this)}
			/>;
		case CommandFieldKind.Checkbox:
			return <input
				type="checkbox"
				name={field.Key}
				className="form-control"
				defaultChecked={field.DefaultValueBoolean}
				onChange={this.onCheckboxChange.bind(this)}
			/>;
		default:
			return unrecognizedValue(field.Kind);
		}
	}
}
