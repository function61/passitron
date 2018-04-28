import * as React from 'react';
import * as ReactDOM from 'react-dom';
import {Router} from 'router';

export default function (appElement: HTMLElement): void {
	function render() {
		ReactDOM.render(
		    <Router initialHash={ document.location.hash } />,
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
