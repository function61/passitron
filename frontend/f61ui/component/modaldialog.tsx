import { Loading } from 'f61ui/component/loading';
import { jsxChildType } from 'f61ui/types';
import { focusRetainer, uniqueDomId } from 'f61ui/utils';
import * as React from 'react';
import * as ReactDOM from 'react-dom';

interface ModalDialogProps {
	title: string;
	onSave: () => void;
	onClose: () => void;
	submitBtnClass: string;
	submitEnabled: boolean;
	loading: boolean;
	children: jsxChildType;
}

export class ModalDialog extends React.Component<ModalDialogProps, {}> {
	private dialogRef: HTMLDivElement | null = null;
	private modalId = 'mdl' + uniqueDomId().toString();

	save() {
		this.props.onSave();
	}

	componentDidMount() {
		// modal showing loses the focus if the focus was already inside the modal content,
		// so we use this hack to retain the focused element
		focusRetainer(() => {
			$(this.dialogRef!).modal('show');
		});

		// we need to let parent know of dialog close, so parent can destroy us,
		// because after dialog has been closed, we are pretty much useless
		$(this.dialogRef!).on('hidden.bs.modal', () => {
			if (this.props.onClose) {
				this.props.onClose();
			}
		});
	}

	render() {
		const labelName = this.modalId + 'Label';

		// abusing modal normal usage by having display:block already applied, because
		// otherwise any autofocused inputs inside this modal won't work
		// (.focus() doesn't work on non-visible inputs)

		const dialogContent = (
			<div
				className="modal"
				style={{ display: 'block' }}
				ref={(input) => {
					this.dialogRef = input;
				}}
				id={this.modalId}
				tabIndex={-1}
				role="dialog"
				aria-labelledby={labelName}>
				<div className="modal-dialog" role="document">
					<div className="modal-content">
						<div className="modal-header">
							<button
								type="button"
								className="close"
								data-dismiss="modal"
								aria-label="Close">
								<span aria-hidden="true">&times;</span>
							</button>
							<h4 className="modal-title" id={labelName}>
								{this.props.title}
							</h4>
						</div>
						<div className="modal-body">{this.props.children}</div>
						<div className="modal-footer">
							{this.props.loading ? <Loading /> : null}
							<button
								type="button"
								className="btn btn-default"
								data-dismiss="modal"
								disabled={this.props.loading}>
								Cancel
							</button>
							<button
								type="button"
								onClick={() => {
									this.save();
								}}
								className={`btn ${this.props.submitBtnClass}`}
								disabled={!this.props.submitEnabled}>
								{this.props.title}
							</button>
						</div>
					</div>
				</div>
			</div>
		);

		// https://reactjs.org/docs/portals.html
		return ReactDOM.createPortal(dialogContent, document.getElementById('appDialogs')!);
	}
}
