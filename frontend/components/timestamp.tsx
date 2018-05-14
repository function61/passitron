import moment = require('moment');
import * as React from 'react';

function relativeDateFormat(dateIso: string): string {
	return moment(dateIso).fromNow();
}

interface TimestampProps {
	ts: string;
}

export class Timestamp extends React.Component<TimestampProps, {}> {
	render() {
		return <span title={this.props.ts}>{relativeDateFormat(this.props.ts)}</span>;
	}
}
