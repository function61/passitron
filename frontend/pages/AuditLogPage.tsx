import * as React from 'react';
import DefaultLayout from 'layouts/DefaultLayout';
import {Breadcrumb} from 'components/breadcrumbtrail';
import {auditLogEntries, defaultErrorHandler} from 'repo';
import {rootFolderName, AuditlogEntry} from 'model';
import {indexLink} from 'links';

interface AuditLogPageState {
	entries: AuditlogEntry[];
}

export default class AuditLogPage extends React.Component<{}, AuditLogPageState> {
	private title = 'Audit log';

	componentDidMount() {
		this.fetchData();
	}

	render() {
		const entryToRow = (entry: AuditlogEntry) => <tr>
			<td>{entry.Timestamp}</td>
			<td>{entry.Message}</td>
		</tr>;

		const rows = this.state && this.state.entries ?
			this.state.entries.map(entryToRow) :
			<tr><td>loading</td></tr>;

		return <DefaultLayout title={this.title} breadcrumbs={this.getBreadcrumbs()}>
			<h1>{this.title}</h1>

			<table>
				<tbody>
					{rows}
				</tbody>
			</table>

		</DefaultLayout>;
	}

	private getBreadcrumbs(): Breadcrumb[] {
		return [
			{url: indexLink(), title: rootFolderName},
			{url: '', title: this.title},
		];
	}

	private fetchData() {
		auditLogEntries().then((entries) => {
			this.setState({
				entries
			});
		}, defaultErrorHandler);
	}
}
