import {CommandDefinition} from 'commandtypes';
import {ModalDialog} from 'components/modaldialog';
import {CommandManager, CommandPagelet} from 'plumbing';
import * as React from 'react';
import {defaultErrorHandler} from 'repo';

interface CommandButtonProps {
	command: CommandDefinition;
}

interface CommandButtonState {
	dialogOpen: boolean;
}

export class CommandButton extends React.Component<CommandButtonProps, CommandButtonState> {
	private cmdManager: CommandManager;

	constructor(props: CommandButtonProps, state: CommandButtonState) {
		super(props, state);

		this.state = { dialogOpen: false };

		this.cmdManager = new CommandManager(this.props.command);
	}

	save() {
		this.cmdManager.execute().then(() => {
			document.location.reload();
		}, defaultErrorHandler);
	}

	render() {
		const commandTitle = this.props.command.title;

		const maybeDialog = this.state.dialogOpen ? <ModalDialog title={commandTitle} onSave={() => { this.save(); }}>
			<CommandPagelet command={this.props.command} onSubmit={() => { this.save(); }} fieldChanged={this.cmdManager.getChangeHandlerBound()} />
		</ModalDialog> : null;

		return <div style={{display: 'inline-block'}}>
			<a className="btn btn-default" onClick={() => { this.setState({ dialogOpen: true }); }}>{commandTitle}</a>

			{ maybeDialog }
		</div>;
	}
}

type EditType = 'add' | 'edit' | 'remove';

interface CommandIconProps {
	command: CommandDefinition;
	type?: EditType;
}

interface CommandIconState {
	dialogOpen: boolean;
}

const typeToIcon: {[key: string]: string} = {
	add: 'glyphicon-plus',
	edit: 'glyphicon-pencil',
	remove: 'glyphicon-remove',
};

export class CommandIcon extends React.Component<CommandIconProps, CommandIconState> {
	private cmdManager: CommandManager;

	constructor(props: CommandIconProps, state: CommandIconState) {
		super(props, state);

		this.state = { dialogOpen: false };

		this.cmdManager = new CommandManager(this.props.command);
	}

	save() {
		this.cmdManager.execute().then(() => {
			document.location.reload();
		}, defaultErrorHandler);
	}

	render() {
		const commandTitle = this.props.command.title;

		const type = this.props.type ? this.props.type : 'edit';
		const icon = typeToIcon[type];

		const maybeDialog = this.state.dialogOpen ? <ModalDialog title={commandTitle} onSave={() => { this.save(); }}>
			<CommandPagelet command={this.props.command} onSubmit={() => { this.save(); }} fieldChanged={this.cmdManager.getChangeHandlerBound()} />
		</ModalDialog> : null;

		return <span className={`glyphicon ${icon} hovericon margin-left`} onClick={() => { this.setState({dialogOpen: true}); }} title={commandTitle}>
			{maybeDialog}
		</span>;
	}
}
