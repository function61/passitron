import * as React from 'react';
import {CommandDefinition} from 'types';
import {defaultErrorHandler} from 'repo';
import {CommandManager, CommandPagelet} from 'plumbing';
import {uniqueDomId} from 'utils';

interface CommandButtonProps {
	command: CommandDefinition;
}

export class CommandButton extends React.Component<CommandButtonProps, {}> {
	private cmdManager: CommandManager;

	constructor(props: CommandButtonProps, state: {}) {
		super(props, state);

		this.cmdManager = new CommandManager(this.props.command);
	}

	save() {
		this.cmdManager.execute().then(() => {
			document.location.reload();
		}, defaultErrorHandler);
	}

	render() {
		const modalId = 'cmdModal' + uniqueDomId().toString();
		const labelName = modalId + 'Label';

		const commandTitle = this.props.command.title;

		return <div style={{display: 'inline-block'}}>
			<a className="btn btn-default" data-toggle="modal" data-target={'#' + modalId}>{commandTitle}</a>

			<div className="modal" id={modalId} tabIndex={-1} role="dialog" aria-labelledby={labelName}>
				<div className="modal-dialog" role="document">
					<div className="modal-content">
						<div className="modal-header">
							<button type="button" className="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>
							<h4 className="modal-title" id={labelName}>{commandTitle}</h4>
						</div>
						<div className="modal-body">
							<CommandPagelet command={this.props.command} updates={this.cmdManager.getChangeHandlerBound()} />
						</div>
						<div className="modal-footer">
							<button type="button" className="btn btn-default" data-dismiss="modal">Close</button>
							<button type="button" onClick={() => this.save()} className="btn btn-primary">Save changes</button>
						</div>
					</div>
				</div>
			</div>
		</div>;
	}
}

interface CommandLinkProps {
	command: CommandDefinition;
}

interface CommandLinkState {
	open: boolean,
}

export class CommandLink extends React.Component<CommandLinkProps, CommandLinkState> {
	private dialogRef: HTMLDivElement | null = null;
	private cmdManager: CommandManager;

	constructor(props: CommandLinkProps, state: CommandLinkState) {
		super(props, state);

		this.cmdManager = new CommandManager(this.props.command);
	}

	componentDidUpdate() {
		if (this.dialogRef) {
			$(this.dialogRef).modal('toggle');
		}
	}

	openDialog() {
		this.setState({ open: true });
	}

	save() {
		this.cmdManager.execute().then(() => {
			document.location.reload();
		}, defaultErrorHandler);
	}

	render() {
		const commandTitle = this.props.command.title;

		if (this.state && this.state.open) {
			const modalId = 'cmdModal' + uniqueDomId().toString();
			const labelName = modalId + 'Label';

			return <div style={{display: 'inline-block'}}>
				<a onClick={() => this.openDialog()}>{commandTitle}</a>

				<div className="modal" ref={(input) => { this.dialogRef = input; }} id={modalId} tabIndex={-1} role="dialog" aria-labelledby={labelName}>
					<div className="modal-dialog" role="document">
						<div className="modal-content">
							<div className="modal-header">
								<button type="button" className="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>
								<h4 className="modal-title" id={labelName}>{commandTitle}</h4>
							</div>
							<div className="modal-body">
								<CommandPagelet command={this.props.command} updates={this.cmdManager.getChangeHandlerBound()} />
							</div>
							<div className="modal-footer">
								<button type="button" className="btn btn-default" data-dismiss="modal">Close</button>
								<button type="button" onClick={() => this.save()} className="btn btn-primary">Save changes</button>
							</div>
						</div>
					</div>
				</div>
			</div>;
		}

		return <a onClick={() => this.openDialog()}>{commandTitle}</a>;
	}
}
