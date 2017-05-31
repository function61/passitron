
function createTable() {
	var table = $('<table class="table table-striped"></table>');
	var thead = $('<thead></thead>').appendTo(table);
	var thTr = $('<tr></tr>').appendTo(thead);
	var tbody = $('<tbody></tbody>').appendTo(table);

	return {
		table: table,

		th: function () {
			return $('<th></th>').appendTo(thTr);
		},

		tr: function () {
			return $('<tr></tr>').appendTo(tbody);
		},

		td: function (tr) {
			return $('<td></td>').appendTo(tr);
		}
	}
}

/*
<ol class="breadcrumb">
  <li><a href="#">Home</a></li>
  <li><a href="#">Library</a></li>
  <li class="active">Data</li>
</ol>
*/
function breadcrumbWidget(items) {
	var breadcrumb = $('<ol class="breadcrumb"></ol>');

	var itemsLen = items.length;
	for (var i = 0; i < itemsLen; ++i) {
		if (i+1 !== itemsLen) {
			var li = $('<li></li>').appendTo(breadcrumb);

			var a = $('<a></a>')
				.attr('href', items[i].href)
				.text(items[i].label)
				.appendTo(li);
		} else {
			$('<li class="active"></li>')
				.text(items[i].label)
				.appendTo(breadcrumb);
		}
	}

	return breadcrumb;
}

function searchWidget(search) {
	return $('<input type="text" class="form-control" placeholder="Search .." />')
		.val(search || '')
		.on('change', function (){
			if (this.value) {
				navigateTo([ 'search', this.value ]);
			} else {
				navigateTo([ 'index' ]);
			}
		});
}

var container;

function layoutInit() {
	document.body.innerHTML = "";

	container = $('<div class="container"></div>').appendTo(document.body);

	var header = $('<div class="header clearfix"></div>').appendTo(container);

	var nav = $('<nav></nav>').appendTo(header);

	var navUl = $('<ul class="nav nav-pills pull-right"></ul>').appendTo(nav);

	function menuItem(href, label) {
		var li = $('<li></li>').appendTo(navUl);

		if (document.location.hash === href) {
			li.addClass('active');
		}

		$('<a></a>').text(label).attr('href', href).appendTo(li);
	}

	menuItem(linkTo([ 'index' ]), 'Home');
	menuItem(linkTo([ 'settings' ]), 'Settings');

	var h3 = $('<h3 class="text-muted"></h3>').appendTo(header);

	$('<a></a>')
		.text('Loq')
		.attr('href', linkTo([ 'index' ]))
		.appendTo(h3);
}

function cc() { // "content container"
	return container;
}

function append(el) {
	$(el).appendTo(container);
}
