import {CommandDefinition, CrudNature} from 'commandtypes';
import {ModalDialog} from 'components/modaldialog';
import {CommandChangesArgs, CommandPagelet, initialCommandState} from 'plumbing';
import * as React from 'react';
import {unrecognizedValue} from 'utils';

interface CommandButtonProps {
	command: CommandDefinition;
}

interface CommandButtonState {
	dialogOpen: boolean;
	cmdState: CommandChangesArgs;
}

export class CommandButton extends React.Component<CommandButtonProps, CommandButtonState> {
	state = { dialogOpen: false, cmdState: initialCommandState() };

	private cmdPagelet: CommandPagelet | null = null;

	save() {
		// FIXME: remove duplication of this code
		this.cmdPagelet!.submit().then(() => {
			document.location.reload();
		}, () => { /* noop */ });
	}

	render() {
		const commandTitle = this.props.command.title;

		const maybeDialog = this.state.dialogOpen ?
			<ModalDialog
				title={commandTitle}
				onSave={() => { this.save(); }}
				loading={this.state.cmdState.processing}
				submitBtnClass={btnClassFromCrudNature(this.props.command.crudNature)}
				submitEnabled={this.state.cmdState.submitEnabled}
			>
				<CommandPagelet
					command={this.props.command}
					onSubmit={() => { this.save(); }}
					onChanges={(cmdState) => { this.setState({ cmdState }); }}
					ref={(el) => { this.cmdPagelet = el; }} />
			</ModalDialog> : null;

		return <div style={{display: 'inline-block'}}>
			<a className="btn btn-default" onClick={() => { this.setState({ dialogOpen: true }); }}>{commandTitle}</a>

			{ maybeDialog }
		</div>;
	}
}

interface CommandIconProps {
	command: CommandDefinition;
}

interface CommandIconState {
	dialogOpen: boolean;
	cmdState: CommandChangesArgs;
}

function commandCrudNatureToIcon(nature: CrudNature): string {
	switch (nature) {
	case CrudNature.create:
		return 'glyphicon-plus';
	case CrudNature.update:
		return 'glyphicon-pencil';
	case CrudNature.delete:
		return 'glyphicon-remove';
	default:
		throw unrecognizedValue(nature);
	}
}

function btnClassFromCrudNature(nature: CrudNature): string {
	switch (nature) {
	case CrudNature.create:
	case CrudNature.update:
		return 'btn-primary';
	case CrudNature.delete:
		return 'btn-danger';
	default:
		throw unrecognizedValue(nature);
	}
}

export class CommandIcon extends React.Component<CommandIconProps, CommandIconState> {
	state = { dialogOpen: false, cmdState: initialCommandState() };

	private cmdPagelet: CommandPagelet | null = null;

	save() {
		// FIXME: remove duplication of this code
		this.cmdPagelet!.submit().then(() => {
			document.location.reload();
		}, () => { /* noop */ });
	}

	render() {
		const commandTitle = this.props.command.title;

		const icon = commandCrudNatureToIcon(this.props.command.crudNature);

		const maybeDialog = this.state.dialogOpen ?
			<ModalDialog
				title={commandTitle}
				onSave={() => { this.save(); }}
				loading={this.state.cmdState.processing}
				submitBtnClass={btnClassFromCrudNature(this.props.command.crudNature)}
				submitEnabled={this.state.cmdState.submitEnabled}
			>
				<CommandPagelet
					command={this.props.command}
					onSubmit={() => { this.save(); }}
					onChanges={(cmdState) => { this.setState({ cmdState }); }}
					ref={(el) => { this.cmdPagelet = el; }} />
			</ModalDialog> : null;

		return <span className={`glyphicon ${icon} hovericon margin-left`} onClick={() => { this.setState({dialogOpen: true}); }} title={commandTitle}>
			{maybeDialog}
		</span>;
	}
}

interface CommandLinkProps {
	command: CommandDefinition;
}

interface CommandLinkState {
	dialogOpen: boolean;
	cmdState: CommandChangesArgs;
}

export class CommandLink extends React.Component<CommandLinkProps, CommandLinkState> {
	state = { dialogOpen: false, cmdState: initialCommandState() };

	private cmdPagelet: CommandPagelet | null = null;

	save() {
		// FIXME: remove duplication of this code
		this.cmdPagelet!.submit().then(() => {
			document.location.reload();
		}, () => { /* noop */ });
	}

	render() {
		const commandTitle = this.props.command.title;

		const maybeDialog = this.state.dialogOpen ?
			<ModalDialog
				title={commandTitle}
				onSave={() => { this.save(); }}
				loading={this.state.cmdState.processing}
				submitBtnClass={btnClassFromCrudNature(this.props.command.crudNature)}
				submitEnabled={this.state.cmdState.submitEnabled}
			>
				<CommandPagelet
					command={this.props.command}
					onSubmit={() => { this.save(); }}
					onChanges={(cmdState) => { this.setState({ cmdState }); }}
					ref={(el) => { this.cmdPagelet = el; }} />
			</ModalDialog> : null;

		return <a className="fauxlink" onClick={() => { this.setState({dialogOpen: true}); }} key={this.props.command.key}>
			{commandTitle}
			{maybeDialog}
		</a>;
	}
}
