import { CommandDefinition, CrudNature } from 'f61ui/commandtypes';
import {
	CommandChangesArgs,
	CommandPagelet,
	initialCommandState,
} from 'f61ui/component/commandpagelet';
import { Loading } from 'f61ui/component/loading';
import { ModalDialog } from 'f61ui/component/modaldialog';
import { unrecognizedValue } from 'f61ui/utils';
import * as React from 'react';

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
		this.cmdPagelet!.submitAndReloadOnSuccess();
	}

	render() {
		const commandTitle = this.props.command.title;

		const maybeDialog = this.state.dialogOpen ? (
			<ModalDialog
				title={commandTitle}
				onClose={() => {
					this.setState({ dialogOpen: false });
				}}
				onSave={() => {
					this.save();
				}}
				loading={this.state.cmdState.processing}
				submitBtnClass={btnClassFromCrudNature(this.props.command.crudNature)}
				submitEnabled={this.state.cmdState.submitEnabled}>
				<CommandPagelet
					command={this.props.command}
					onSubmit={() => {
						this.save();
					}}
					onChanges={(cmdState) => {
						this.setState({ cmdState });
					}}
					ref={(el) => {
						this.cmdPagelet = el;
					}}
				/>
			</ModalDialog>
		) : null;

		return (
			<div style={{ display: 'inline-block' }}>
				<a
					className="btn btn-default"
					onClick={() => {
						this.setState({ dialogOpen: true });
					}}>
					{commandTitle}
				</a>

				{maybeDialog}
			</div>
		);
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
		this.cmdPagelet!.submitAndReloadOnSuccess();
	}

	render() {
		const commandTitle = this.props.command.title;

		const icon = commandCrudNatureToIcon(this.props.command.crudNature);

		const maybeDialog = this.state.dialogOpen ? (
			<ModalDialog
				title={commandTitle}
				onClose={() => {
					this.setState({ dialogOpen: false });
				}}
				onSave={() => {
					this.save();
				}}
				loading={this.state.cmdState.processing}
				submitBtnClass={btnClassFromCrudNature(this.props.command.crudNature)}
				submitEnabled={this.state.cmdState.submitEnabled}>
				<CommandPagelet
					command={this.props.command}
					onSubmit={() => {
						this.save();
					}}
					onChanges={(cmdState) => {
						this.setState({ cmdState });
					}}
					ref={(el) => {
						this.cmdPagelet = el;
					}}
				/>
			</ModalDialog>
		) : null;

		return (
			<span
				className={`glyphicon ${icon} hovericon margin-left`}
				onClick={() => {
					this.setState({ dialogOpen: true });
				}}
				title={commandTitle}>
				{maybeDialog}
			</span>
		);
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
		this.cmdPagelet!.submitAndReloadOnSuccess();
	}

	render() {
		const commandTitle = this.props.command.title;

		const maybeDialog = this.state.dialogOpen ? (
			<ModalDialog
				title={commandTitle}
				onClose={() => {
					this.setState({ dialogOpen: false });
				}}
				onSave={() => {
					this.save();
				}}
				loading={this.state.cmdState.processing}
				submitBtnClass={btnClassFromCrudNature(this.props.command.crudNature)}
				submitEnabled={this.state.cmdState.submitEnabled}>
				<CommandPagelet
					command={this.props.command}
					onSubmit={() => {
						this.save();
					}}
					onChanges={(cmdState) => {
						this.setState({ cmdState });
					}}
					ref={(el) => {
						this.cmdPagelet = el;
					}}
				/>
			</ModalDialog>
		) : null;

		return (
			<a
				className="fauxlink"
				onClick={() => {
					this.setState({ dialogOpen: true });
				}}
				key={this.props.command.key}>
				{commandTitle}
				{maybeDialog}
			</a>
		);
	}
}

interface CommandInlineFormProps {
	command: CommandDefinition;
}

interface CommandInlineFormState {
	cmdState?: CommandChangesArgs;
}

export class CommandInlineForm extends React.Component<
	CommandInlineFormProps,
	CommandInlineFormState
> {
	public state: CommandInlineFormState = {};
	private cmdPagelet: CommandPagelet | null = null;

	render() {
		const submitEnabled = this.state.cmdState && this.state.cmdState.submitEnabled;
		const maybeLoading =
			this.state.cmdState && this.state.cmdState.processing ? <Loading /> : '';

		return (
			<div>
				<CommandPagelet
					command={this.props.command}
					onSubmit={() => {
						this.save();
					}}
					onChanges={(cmdState) => {
						this.setState({ cmdState });
					}}
					ref={(el) => {
						this.cmdPagelet = el;
					}}
				/>

				<button
					className="btn btn-primary"
					onClick={() => {
						this.save();
					}}
					disabled={!submitEnabled}>
					{this.props.command.title}
				</button>

				{maybeLoading}
			</div>
		);
	}

	save() {
		this.cmdPagelet!.submitAndReloadOnSuccess();
	}
}
