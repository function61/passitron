import { Breadcrumb } from 'f61ui/component/breadcrumbtrail';
import { Loading } from 'f61ui/component/loading';
import { Timestamp } from 'f61ui/component/timestamp';
import { defaultErrorHandler } from 'f61ui/errors';
import { auditLogEntries } from 'generated/apitypes_endpoints';
import { AuditlogEntry } from 'generated/apitypes_types';
import { RootFolderName } from 'generated/domain_types';
import { AppDefaultLayout } from 'layout/appdefaultlayout';
import * as React from 'react';
import { indexUrl } from 'generated/apitypes_uiroutes';

interface AuditLogPageState {
	entries: AuditlogEntry[];
}

export default class AuditLogPage extends React.Component<{}, AuditLogPageState> {
	private title = 'Audit log';

	componentDidMount() {
		this.fetchData();
	}

	render() {
		const entryToRow = (entry: AuditlogEntry, idx: number) => (
			<tr key={idx}>
				<td>
					<Timestamp ts={entry.Timestamp} />
				</td>
				<td>{entry.UserId}</td>
				<td>{entry.Message}</td>
			</tr>
		);

		const rows =
			this.state && this.state.entries ? (
				this.state.entries.map(entryToRow)
			) : (
				<tr>
					<td>
						<Loading />
					</td>
				</tr>
			);

		return (
			<AppDefaultLayout title={this.title} breadcrumbs={this.getBreadcrumbs()}>
				<h1>{this.title}</h1>

				<table className="table table-striped">
					<tbody>{rows}</tbody>
				</table>
			</AppDefaultLayout>
		);
	}

	private getBreadcrumbs(): Breadcrumb[] {
		return [
			{ url: indexUrl(), title: RootFolderName },
			{ url: '', title: this.title },
		];
	}

	private fetchData() {
		auditLogEntries().then((entries) => {
			this.setState({
				entries,
			});
		}, defaultErrorHandler);
	}
}
