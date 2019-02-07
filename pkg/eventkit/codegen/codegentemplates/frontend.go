package codegentemplates

const FrontendDatatypes = `// tslint:disable
// WARNING: generated file

{{range .TypeDependencyModuleIds}}
import * as {{.}} from '{{$.FrontendModulePrefix}}{{.}}_types';{{end}}
{{if .TypesDependOnTime}}import {datetimeRFC3339} from 'f61ui/types';
{{end}}

{{range .StringEnums}}
export enum {{.Name}} {
{{range .Members}}
	{{.Key}} = '{{.GoValue}}',{{end}}
}
{{end}}
{{range .ApplicationTypes.StringConsts}}
export const {{.Key}} = '{{.Value}}';{{end}}
{{range .ApplicationTypes.Types}}
{{.AsTypeScriptCode}}
{{end}}
`

const FrontendRestEndpoints = `// tslint:disable
// WARNING: generated file

// WHY: wouldn't make sense complicating code generation to check
// if we need template string or not in path string

import { {{range .ApplicationTypes.EndpointsProducesAndConsumesTypescriptTypes}}
	{{.}},{{end}}
} from '{{$.FrontendModulePrefix}}{{.ModuleId}}_types';
import {getJson, postJson} from 'f61ui/httputil';

{{range .ApplicationTypes.Endpoints}}
// {{.Path}}
export function {{.Name}}({{.TypescriptArgs}}) {
	return {{if .Consumes}}postJson<{{if .Consumes}}{{.Consumes.AsTypeScriptType}}{{else}}void{{end}}, {{if .Produces}}{{.Produces.AsTypeScriptType}}{{else}}void{{end}}>{{else}}getJson<{{if .Produces}}{{.Produces.AsTypeScriptType}}{{else}}void{{end}}>{{end}}(` + "`{{.TypescriptPath}}`" + `{{if .Consumes}}, body{{end}});
}
{{if not .Consumes}}
export function {{.Name}}Url({{.TypescriptArgs}}): string {
	return ` + "`{{.TypescriptPath}}`" + `;
}{{end}}
{{end}}
`

const FrontendCommandDefinitions = `// tslint:disable
// WARNING: generated file

import {CommandDefinition, CommandFieldKind, CrudNature} from 'f61ui/commandtypes';

{{range .CommandSpecs}}
export function {{.AsGoStructName}}({{.CtorArgsForTypeScript}}): CommandDefinition {
	return {
		key: '{{.Command}}',{{if .AdditionalConfirmation}}
		additional_confirmation: '{{.AdditionalConfirmation}}',
{{end}}		title: '{{.Title}}',
		crudNature: CrudNature.{{.CrudNature}},
		fields: [
{{.FieldsForTypeScript}}
		],
	};
}
{{end}}
`

const FrontendVersion = `// tslint:disable
// WARNING: generated file

export const version = '{{.Version}}';
`
