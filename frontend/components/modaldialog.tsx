import {Loading} from 'components/loading';
import * as React from 'react';
import * as ReactDOM from 'react-dom';
import {uniqueDomId} from 'utils';

interface ModalDialogProps {
	title: string;
	onSave: () => void;
	submitBtnClass: string;
	submitEnabled: boolean;
	loading: boolean;
	children: JSX.Element | JSX.Element[];
}

export class ModalDialog extends React.Component<ModalDialogProps, {}> {
	private dialogRef: HTMLDivElement | null = null;

	save() {
		this.props.onSave();
	}

	componentDidMount() {
		$(this.dialogRef!).modal('toggle');
	}

	render() {
		const modalId = 'cmdModal' + uniqueDomId().toString();
		const labelName = modalId + 'Label';

		const dialogContent = <div className="modal" ref={(input) => { this.dialogRef = input; }} id={modalId} tabIndex={-1} role="dialog" aria-labelledby={labelName}>
			<div className="modal-dialog" role="document">
				<div className="modal-content">
					<div className="modal-header">
						<button type="button" className="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>
						<h4 className="modal-title" id={labelName}>{this.props.title}</h4>
					</div>
					<div className="modal-body">
						{this.props.children}
					</div>
					<div className="modal-footer">
						{this.props.loading ? <Loading /> : null}
						<button type="button" className="btn btn-default" data-dismiss="modal" disabled={this.props.loading}>Cancel</button>
						<button type="button" onClick={() => { this.save(); }} className={`btn ${this.props.submitBtnClass}`} disabled={!this.props.submitEnabled}>{this.props.title}</button>
					</div>
				</div>
			</div>
		</div>;

		// https://reactjs.org/docs/portals.html
		return ReactDOM.createPortal(dialogContent, document.getElementById('appDialogs')!);
	}
}
