
function navigateTo(components) {
	document.location.assign(linkTo(components));
}

function linkTo(components) {
	return '#' + components.join('/');
}

var routes = {}; // filled from other modules

function renderPage(components) {
	layoutInit();

	var page = components[0];

	if (!(page in routes)) {
		alert('unknown page');
		return;
	}

	routes[page].call(this, components);
}

function softReload() {
	hashChanged();
}

function hashChanged() {
	if (!document.location.hash) {
		navigateTo([ 'index' ]);
		return;
	}

	// '#index' => 'index'
	renderPage(document.location.hash.substr(1).split('/'));
}

$(document).ready(function (){
	// uses event delegation, so nodes created later are ok
	new Clipboard('[data-clipboard-target]');

	$(window).on('hashchange', hashChanged);

	initCommandArchitecture();

	hashChanged();
});
