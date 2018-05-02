import * as React from 'react';
import * as ReactDOM from 'react-dom';
import {App} from 'app';

export default function (appElement: HTMLElement): void {
	function render() {
		ReactDOM.render(
		    <App initialHash={ document.location.hash } />,
		    appElement);
	}

	/*
	secretStore.subscribe(() => {
		console.log('change in data');
		render();
	});
	*/

	render();
}
