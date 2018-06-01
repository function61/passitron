import * as React from 'react';
import {relativeDateFormat} from 'utils';

interface TimestampProps {
	ts: string;
}

export class Timestamp extends React.Component<TimestampProps, {}> {
	render() {
		return <span title={this.props.ts}>{relativeDateFormat(this.props.ts)}</span>;
	}
}
