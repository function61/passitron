import * as React from 'react';

interface DropdownProps {
	children: JSX.Element[];
}

export class Dropdown extends React.Component<DropdownProps, {}> {
	render() {
		const items = this.props.children.map((child, idx) => {
			return <li key={idx}>{child}</li>;
		});

		return <div className="btn-group">
			<button type="button" className="btn btn-default dropdown-toggle" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
				<span className="caret"></span>
			</button>
			<ul className="dropdown-menu">
				{ items }
			</ul>
		</div>;
	}
}
