import { datetimeRFC3339 } from 'f61ui/types';
import { relativeDateFormat } from 'f61ui/utils';
import * as React from 'react';

interface TimestampProps {
	ts: datetimeRFC3339;
}

export class Timestamp extends React.Component<TimestampProps, {}> {
	render() {
		return <span title={this.props.ts}>{relativeDateFormat(this.props.ts)}</span>;
	}
}
