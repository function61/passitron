import { globalConfig } from 'f61ui/globalconfig';
import * as React from 'react';

export class Loading extends React.Component<{}, {}> {
	render() {
		return <img src={globalConfig().assetsDir + '/loading.gif'} alt="Loading" />;
	}
}
