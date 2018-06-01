import {CommandDefinition, CommandField, CommandFieldKind} from 'commandtypes';
import {defaultErrorHandler} from 'generated/restapi';
import {httpMustBeOk} from 'httputil';
import * as React from 'react';
import {unrecognizedValue} from 'utils';

export type CommandSubmitListener = () => void;

interface CommandPageletProps {
	command: CommandDefinition;
	onSubmit: CommandSubmitListener;
}

interface CommandPageletState {
	values: {[key: string]: any};
	validationStatuses: {[key: string]: boolean};
}

export class CommandPagelet extends React.Component<CommandPageletProps, CommandPageletState> {
	constructor(props: CommandPageletProps) {
		super(props);

		const state: CommandPageletState = { values: {}, validationStatuses: {} };

		// copy default values to values, because they are only updated on
		// "onChange" event, and thus if user doesn't change them, they wouldn't get filled
		this.props.command.fields.forEach((field) => {
			switch (field.Kind) {
			case CommandFieldKind.Integer:
				state.values[field.Key] = null;
				break;
			case CommandFieldKind.Password:
			case CommandFieldKind.Text:
			case CommandFieldKind.Multiline:
				state.values[field.Key] = field.DefaultValueString;
				break;
			case CommandFieldKind.Checkbox:
				state.values[field.Key] = field.DefaultValueBoolean;
				break;
			default:
				unrecognizedValue(field.Kind);
			}

			state.validationStatuses[field.Key] = this.validate(field, state.values[field.Key]);
		});

		this.state = state;
	}

	validate(field: CommandField, value: any): boolean {
		if (field.Required && (value === undefined || value === null || value === '')) {
			return false;
		}

		return true; // if no errors found
	}

	render() {
		const fieldGroups = this.props.command.fields.map((field) => {
			const input = this.createInput(field);

			const valid = this.state.validationStatuses[field.Key];

			const validationFailedClass = valid ? '' : 'has-error';

			return <div className={`form-group ${validationFailedClass}`} key={field.Key}>
				<label>{field.Key} {field.Required ? '*' : ''}</label>

				{input}
			</div>;
		});

		return <form onSubmit={(e) => { this.onInternalEnterSubmit(e); }}>
			{fieldGroups}
		</form>;
	}

	execute(): Promise<void> {
		const bodyToPost = JSON.stringify(this.state.values);

		return fetch(`/command/${this.props.command.key}`, {
			method: 'POST',
			headers: {
				'Accept': 'application/json',
				'Content-Type': 'application/json',
			},
			body: bodyToPost,
		})
			.then(httpMustBeOk)
			.then((res) => res.json())
			.then((_) => {
				return;
			});
	}

	// official submit, which should trigger validation
	submit(): Promise<void> {
		const someInvalid = Object.keys(this.state.validationStatuses).some((key) => !this.state.validationStatuses[key]);

		if (someInvalid) {
			return Promise.reject(new Error('Invalid form data'));
		}

		return this.execute();
	}

	onInternalEnterSubmit(e: React.FormEvent<HTMLFormElement>) {
		e.preventDefault(); // prevent browser-based submit

		this.submit().then(() => {
			document.location.reload();
		}, defaultErrorHandler);
	}

	private updateFieldValue(key: string, value: any) {
		const field = this.fieldByKey(key);

		this.state.values[key] = value;
		this.state.validationStatuses[field.Key] = this.validate(field, value);
		this.setState(this.state);
	}

	private onInputChange(e: React.FormEvent<HTMLInputElement>) {
		this.updateFieldValue(e.currentTarget.name, e.currentTarget.value);
	}

	private onIntegerInputChange(e: React.FormEvent<HTMLInputElement>) {
		if (e.currentTarget.value === '') {
			this.updateFieldValue(e.currentTarget.name, null);
		} else {
			this.updateFieldValue(e.currentTarget.name, +e.currentTarget.value);
		}
	}

	private onCheckboxChange(e: React.FormEvent<HTMLInputElement>) {
		this.updateFieldValue(e.currentTarget.name, e.currentTarget.checked);
	}

	private onTextareaChange(e: React.FormEvent<HTMLTextAreaElement>) {
		this.updateFieldValue(e.currentTarget.name, e.currentTarget.value);
	}

	private createInput(field: CommandField): JSX.Element {
		switch (field.Kind) {
		case CommandFieldKind.Password:
			return <input
				type="password"
				className="form-control"
				name={field.Key}
				required={field.Required}
				value={this.state.values[field.Key]}
				onChange={this.onInputChange.bind(this)}
			/>;
		case CommandFieldKind.Text:
			return <input
				type="text"
				className="form-control"
				name={field.Key}
				required={field.Required}
				value={this.state.values[field.Key]}
				onChange={this.onInputChange.bind(this)}
			/>;
		case CommandFieldKind.Multiline:
			return <textarea
				name={field.Key}
				required={field.Required}
				className="form-control"
				rows={7}
				value={this.state.values[field.Key]}
				onChange={this.onTextareaChange.bind(this)}
			/>;
		case CommandFieldKind.Integer:
			return <input
				type="number"
				className="form-control"
				name={field.Key}
				required={field.Required}
				onChange={this.onIntegerInputChange.bind(this)}
			/>;
		case CommandFieldKind.Checkbox:
			return <input
				type="checkbox"
				name={field.Key}
				className="form-control"
				checked={this.state.values[field.Key]}
				onChange={this.onCheckboxChange.bind(this)}
			/>;
		default:
			return unrecognizedValue(field.Kind);
		}
	}

	private fieldByKey(key: string): CommandField {
		const matches = this.props.command.fields.filter((field) => field.Key === key);

		if (matches.length !== 1) {
			throw new Error(`Field by key ${key} not found`);
		}

		return matches[0];
	}
}
