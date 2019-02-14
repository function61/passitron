// nominal type
enum datetimeRFC3339_ {}
export type datetimeRFC3339 = datetimeRFC3339_ & string;

type jsxChildItem = JSX.Element | string;

// FFS this is the problem with JS culture.. if this can be a list, then why not a
// single-item child could be a list with n=1 instead of having this frankenstein type??!
export type jsxChildType = jsxChildItem | jsxChildItem[];

export interface StructuredErrorResponse {
	error_code: string;
	error_description: string;
}
