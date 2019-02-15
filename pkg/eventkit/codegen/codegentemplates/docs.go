package codegentemplates

const DocsEvents = `Each and every *Event* additionally has the following common meta data:

- Timestamp, in UTC, the event was raised on
- ID of the user that caused that event


{{range .DomainSpecs.Events}}
{{.Event}}
-------

{{if .Changelog}}
Changelog:
{{range .Changelog}}
- {{.}}{{end}}
{{end}}

| key | type | notes |
|-----|------|-------|
{{range .Fields}}| {{.Key}} | {{.Type.Name}} | {{.Notes}} |
{{end}}
{{end}}
`

const DocsTypes = `{{if .ApplicationTypes.StringConsts}}
Constants
---------

| const | value |
|-------|-------|
{{range .ApplicationTypes.StringConsts}}| {{.Key}} | {{.Value}} |
{{end}}
{{end}}

{{range .ApplicationTypes.Enums}}
enum {{.Name}}
---------

{{range .StringMembers}}
- {{.}}{{end}}
{{end}}

{{range .ApplicationTypes.Types}}
{{.Name}}
---------

` + "```" + `
{{.AsTypeScriptCode}}
` + "```" + `
{{end}}
`

const DocsCommands = `Overview
--------

| Endpoint | Middleware | Title |
|----------|------------|-------| {{range .CommandSpecs}}
| POST /command/{{.Command}} | {{.MiddlewareChain}} | {{.Title}} | {{end}}

{{range .CommandSpecs}}
{{.Command}}
------------

| Field | Type | Required | Notes |
|-------|------|----------|-------|
{{range .Fields}}| {{.Key}} | {{.Type}} | {{not .Optional}} | {{.Help}} |
{{end}}
{{end}}
`

const DocsRestEndpoints = `Overview
========

| Path | Middleware | Input | Output | Notes |
|------|------------|-------|--------|-------|
{{range .ApplicationTypes.Endpoints}}| {{.HttpMethod}} {{.Path}} | {{.MiddlewareChain}} | {{if .Consumes}}{{.Consumes.AsTypeScriptType}}{{end}} | {{if .Produces}}{{.Produces.AsTypeScriptType}}{{end}} | {{.Description}} |
{{end}}

{{range .ApplicationTypes.Endpoints}}
{{.HttpMethod}} {{.Path}}
=========================

| Detail           |                                                       |
|------------------|-------------------------------------------------------|
| Middleware chain | {{.MiddlewareChain}}                                  |
| Consumes         | {{if .Consumes}}{{.Consumes.AsTypeScriptType}}{{end}} |
| Produces         | {{if .Produces}}{{.Produces.AsTypeScriptType}}{{end}} |
| Description      | {{.Description}}                                      |

{{end}}
`
