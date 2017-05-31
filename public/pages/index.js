
function credsWidget(matches) {
	var credsTable = createTable();

	credsTable.th().text('Title');
	credsTable.th().text('Username');

	for (var i = 0; i < matches.length; ++i) {
		var tr = credsTable.tr();

		var titleTd = credsTable.td(tr);
		$('<a></a>')
			.attr('href', linkTo([ 'credview', matches[i].Id ]))
			.text(matches[i].Title)
			.appendTo(titleTd);

		credsTable.td(tr).text(matches[i].Username);
	}

	return credsTable.table;
}

routes.index = function(args) {
	var folderId = args[1] || 'root';


	byFolder(folderId).then(function (resp) {
		searchWidget(null).appendTo(cc());

		var bcItems = resp.ParentFolders.reverse().map(function (item){
			return {
				href: linkTo([ 'index', item.Id ]),
				label: item.Name
			};
		});

		bcItems.push({
			href: '', // don't need one - last item is not linked to
			label: resp.Folder.Name
		});

		breadcrumbWidget(bcItems).appendTo(cc());

		for (var i = 0; i < resp.SubFolders.length; ++i) {
			var sub = resp.SubFolders[i];

			var folderIcon = $('<span class="glyphicon glyphicon-folder-open"></span>');

			$('<a class="btn btn-default"></a>')
				.attr('href', linkTo([ 'index', sub.Id ]))
				.append([ folderIcon, ' &nbsp;', sub.Name ])
				// .text(sub.Name)
				.appendTo(cc());
		}

		credsWidget(resp.Secrets).appendTo(cc());

		var secretCreateBtn = $('<button class="btn btn-default"></button>')
			.text('+ Secret')
			.appendTo(cc());

		attachCommand(secretCreateBtn, {
			cmd: 'SecretCreateRequest',
			prefill: {
				FolderId: folderId
			} });

		var folderCreateBtn = $('<button class="btn btn-default"></button>')
			.text('+ Folder')
			.appendTo(cc());

		attachCommand(folderCreateBtn, {
			cmd: 'FolderCreateRequest',
			prefill: {
				ParentId: folderId
			} });
	});
}
