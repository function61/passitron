import { defaultErrorHandler } from 'backenderrors';
import { Breadcrumb } from 'f61ui/components/breadcrumbtrail';
import { Loading } from 'f61ui/components/loading';
import { Timestamp } from 'f61ui/components/timestamp';
import { AuditlogEntry } from 'generated/apitypes';
import { RootFolderName } from 'generated/domain';
import { auditLogEntries } from 'generated/restapi';
import { AppDefaultLayout } from 'layout/appdefaultlayout';
import * as React from 'react';
import { indexRoute } from 'routes';

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
			{ url: indexRoute.buildUrl({}), title: RootFolderName },
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
