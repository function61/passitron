import * as React from 'react';
import { datetimeRFC3339 } from 'types';
import { relativeDateFormat } from 'utils';

interface TimestampProps {
	ts: datetimeRFC3339;
}

export class Timestamp extends React.Component<TimestampProps, {}> {
	render() {
		return <span title={this.props.ts}>{relativeDateFormat(this.props.ts)}</span>;
	}
}
