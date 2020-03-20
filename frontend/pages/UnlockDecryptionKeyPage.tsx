import { WarningAlert } from 'f61ui/component/alerts';
import { Panel } from 'f61ui/component/bootstrap';
import { Breadcrumb } from 'f61ui/component/breadcrumbtrail';
import { CommandInlineForm } from 'f61ui/component/CommandButton';
import { UserUnlockDecryptionKey } from 'generated/apitypes_commands';
import { RootFolderName } from 'generated/domain_types';
import { AppDefaultLayout } from 'layout/appdefaultlayout';
import * as React from 'react';
import { indexUrl } from 'generated/apitypes_uiroutes';

interface UnlockDecrypionKeyPageProps {
	redirect: string;
}

export default class UnlockDecrypionKeyPage extends React.Component<
	UnlockDecrypionKeyPageProps,
	{}
> {
	private title = 'Unlock decryption key';

	render() {
		return (
			<AppDefaultLayout title={this.title} breadcrumbs={this.getBreadcrumbs()}>
				<Panel heading={this.title}>
					<WarningAlert>
						Your decryption key is locked - enter its password to be able to access your
						secrets.
					</WarningAlert>

					<CommandInlineForm
						command={UserUnlockDecryptionKey({ redirect: () => this.props.redirect })}
					/>
				</Panel>
			</AppDefaultLayout>
		);
	}

	private getBreadcrumbs(): Breadcrumb[] {
		return [
			{ url: indexUrl(), title: RootFolderName },
			{ url: '', title: this.title },
		];
	}
}
