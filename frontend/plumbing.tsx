import {CommandDefinition, CommandField, CommandFieldKind} from 'commandtypes';
import {coerceToStructuredErrorResponse, defaultErrorHandler, handleDatabaseSealed, StructuredErrorResponse} from 'generated/restapi';
import {httpMustBeOk} from 'httputil';
import * as React from 'react';
import {unrecognizedValue} from 'utils';

export type CommandSubmitListener = () => void;

export interface CommandChangesArgs {
	submitEnabled: boolean;
	// server is currently processing this request?
	processing: boolean;
}

export function initialCommandState(): CommandChangesArgs {
	return { submitEnabled: false, processing: false };
}

export type CommandChangesListener = (cmdState: CommandChangesArgs) => void;

interface CommandPageletProps {
	command: CommandDefinition;
	onChanges: CommandChangesListener;
	onSubmit: CommandSubmitListener;
}

interface CommandPageletState {
	values: {[key: string]: any};
	validationStatuses: {[key: string]: boolean};
	submitError: string;
	fieldsThatWerePrefilled: {[key: string]: boolean};
}

export class CommandPagelet extends React.Component<CommandPageletProps, CommandPageletState> {
	constructor(props: CommandPageletProps) {
		super(props);

		const state: CommandPageletState = {
			values: {},
			validationStatuses: {},
			submitError: '',
			fieldsThatWerePrefilled: {},
		};

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
				if (field.DefaultValueString) {
					state.fieldsThatWerePrefilled[field.Key] = true;
				}
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

		// so that initial validation state is used - otherwise only the
		// first onchange would yield in current state
		this.broadcastChanges();
	}

	render() {
		const shouldShow = (field: CommandField) =>
			!field.HideIfDefaultValue || !(field.Key in this.state.fieldsThatWerePrefilled);

		const fieldGroups = this.props.command.fields.filter(shouldShow).map((field) => {
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

			{this.state.submitError ? <p className="bg-danger">{this.state.submitError}</p> : null}
		</form>;
	}

	// official submit, which should trigger validation
	submit(): Promise<void> {
		// disable submit button while server is processing
		this.props.onChanges({
			processing: true,
			submitEnabled: false,
		});

		this.setState({ submitError: '' });

		const execResult = this.execute();

		// whether fulfilled or rejected, return submitEnabled
		// state back to what it should be
		execResult.then(() => {
			this.broadcastChanges();
		}, (err: Error | StructuredErrorResponse) => {
			const ser = coerceToStructuredErrorResponse(err);
			if (handleDatabaseSealed(ser)) {
				return;
			}

			this.setState({ submitError: `${ser.error_code}: ${ser.error_description}` });

			this.broadcastChanges();
		});

		return execResult;
	}

	private validate(field: CommandField, value: any): boolean {
		if (field.Required && (value === undefined || value === null || value === '')) {
			return false;
		}

		return true; // if no errors found
	}

	private execute(): Promise<void> {
		if (!this.isEverythingValid()) {
			return Promise.reject(new Error('Invalid form data'));
		}

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
			.then((res) => res.json());
	}

	private broadcastChanges() {
		this.props.onChanges({
			submitEnabled: this.isEverythingValid(),
			processing: false,
		});
	}

	private onInternalEnterSubmit(e: React.FormEvent<HTMLFormElement>) {
		e.preventDefault(); // prevent browser-based submit

		this.submit().then(() => {
			document.location.reload();
		}, defaultErrorHandler);
	}

	private isEverythingValid() {
		return Object.keys(this.state.validationStatuses).every((key) => this.state.validationStatuses[key]);
	}

	private updateFieldValue(key: string, value: any) {
		const field = this.fieldByKey(key);

		this.state.values[key] = value;
		this.state.validationStatuses[field.Key] = this.validate(field, value);

		this.setState(this.state);

		// TODO: this sets processing: false. this should not be
		// done while the server is actually processing
		this.broadcastChanges();
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
