import {toYyyymmdd, fromYyyymmdd} from "./utilities/date";
import {safeParseNum} from "./utilities/num";

/* Preface: At the time of writing, there were no two-way (build and parse) typesafe routing libraries available for typescript,
 * So I wrote my own. There is some type-level programming needed before it's actually type safe, so
 * it's not the easiest code to read.
 * Typesafe in this context means: Declare your routes once, run them through the library, get safely typed
 * generator and parser functions for your routes.
 */

/* TODO
 * Hardcoded usage of "#/" OK?
 * A url that matches a route but has other stuff in the URL aswell still matches the route. Is this OK?
 */

/** This interface holds all the types of values that can be encoded in a URL string */
export interface SerializableTypeNameToType {
  "string": string;
  "string | null": string | null;
  "date": Date;
  "date | null": Date | null;
  "number": number;
  "number | null": number | null;
  "boolean": boolean;
  "boolean | null": boolean | null;
}
export type SerializableTypeName = keyof SerializableTypeNameToType;

/** A type describing the map of parameter names to their types for a route */
export type RouteParamsSpec = {[paramName: string]: SerializableTypeName};

/** A type-level helper function to translate from:
  *   a RouteParamsSpec, ie. a map from parameter names to types in string format
  *   to a map from the same parameter names to the actual types
 */
export type SpecToType<S extends RouteParamsSpec> = {[K in keyof S]: SerializableTypeNameToType[S[K]]}

function isNullable(t: SerializableTypeName): boolean {
  if (t === "string | null" || t === "number | null" || t === "date | null" || t === "boolean | null") {
    return true;
  } else {
    return false;
  }
}

/** Internal function for serializing URL parts based on Route definitions */
function serializeRouteParam<P extends keyof SerializableTypeNameToType>(t: P, val: SerializableTypeNameToType[P]): string {
  // Stupid casts... But typescript analysis can't figure out SerializableTypeNameToType[P] correctly inside the if statement
  if (t === "string" || t === "string | null") {
    return val as string;
  } else if (t === "number" || t === "number | null") {
    return (val as number).toString();
  } else if (t === "date" || t === "date | null") {
    return toYyyymmdd(val as Date);
  } else if (t === "boolean" || t === "boolean | null") {
    return (val ? "true" : "false");
  } else {
    throw new Error("There is a case statement missing in serializeRouteParam for type " + t);
  }
}

/** Internal function for deserializing URL parts based on Route definitions */
function deserializeRouteParam<P extends keyof SerializableTypeNameToType>(t: P, val: string): SerializableTypeNameToType[P] | null {
  // Nulls should NEVER come into this function
  if (t === "string" || t === "string | null") {
    return val;
  } else if (t === "number" || t === "number | null") {
    return safeParseNum(val);
  } else if (t === "date" || t === "date | null") {
    return fromYyyymmdd(val);
  } else if (t === "boolean" || t === "boolean | null") {
    return val === "true" ? true : (val === "false" ? false : null);
  } else {
    throw new Error("There is a case statement missing in deserializeRouteParam for type " + t);
  }
}

/* TODO: It would be nicer if we could use Route<{myString: string}>, but I haven't
 *   found a way to make that work
 */
/** A route is a pair of functions:
  *   to match a URL against a route specification
  *   and build a URL from data confirming to that specification
  */
export type Route<S extends RouteParamsSpec> = {
  matchUrl: (hash: string) => SpecToType<S> | null;
  buildUrl: (params: SpecToType<S>) => string;
}

/** Make a Route from a name and a specification of its properties */
export function makeRoute<S extends RouteParamsSpec>(routeName: string, routeParams: S): Route<S> {
  return {
    matchUrl: function(hash: string): SpecToType<S> | null {
      const startStr = "#/" + routeName;
      if (!startsWith(hash, startStr)) {
        return null;
      }
      let params: Partial<SpecToType<S>> = {};
      Object.keys(routeParams).forEach(function<K extends keyof S>(key: K){
        const val: SerializableTypeName = routeParams[key];
        let found = findInHash(hash, key as string);
        if (found === null) {
          if (isNullable(val)) {
            params[key] = null;
            return;
          } else {
            throw new Error(`Parsing route ${routeName}: Couldn't find route param ${key}`);
          }
        } else {
          let deser = deserializeRouteParam(val, decodeURIComponent(found));
          if (deser === null) {
            throw new Error(`Parsing route ${routeName}: Couldn't deserialize route param ${key} from value: ${found}`);
          } else {
            params[key] = deser;
          }
        }
      });
      return params as SpecToType<S>; // Can this cast be avoided?
    },
    buildUrl: function(params: SpecToType<S>): string {
      return "#/" + routeName + Object.keys(params).reduce(function<K extends keyof S>(acc: string, key: K){
        const val: SerializableTypeNameToType[S[K]] = params[key];
        // Nullable types -> Don't add the key nor the value to the URL
        if (val === null) {
          return acc;
        } else {
          return acc
            + "/" +  (key as string).toLowerCase()
            // Encoding so we don't break on strings with '/', ...
            + "/" + encodeURIComponent(
              serializeRouteParam(routeParams[key], val));
        }
      }, "");
    }
  };
}

export interface Router<T> {
  match: (hash: string) => T | null;
  registerRoute<S extends RouteParamsSpec>(matcher: Route<S>, handler: (params: SpecToType<S>) => T): Router<T>;
}

/* TODO newRouter vs mkNextRouter contains some duplication.
 *      + newrouter vs registerRoute is a bit of a stupid API
 *   Can this be done better while keeping the T in Router<T> typesafe?
 */
export function makeRouter<T, S extends RouteParamsSpec>(matcher: Route<S>, handler: (params: SpecToType<S>) => T): Router<T> {
  return {
    match: function(hash: string){
      const matchedParams = matcher.matchUrl(hash);
      return matchedParams === null ? null : handler(matchedParams);
    },
    registerRoute<U extends RouteParamsSpec>(matcher: Route<U>, handler: (params: SpecToType<U>) => T): Router<T> {
      return mkNextRouter(this, matcher, handler);
    }
  };
}

/* TODO Never return null from a handler!
 *   Is this an issue of using null in Route.matchUrl instead of Maybe<X>?
 */
function mkNextRouter<T, S extends RouteParamsSpec>(prevRouter: Router<T>, matcher: Route<S>, handler: (params: SpecToType<S>) => T): Router<T> {
  return {
    match: function(hash: string){
      const resultOfPrevious = prevRouter.match(hash);
      if (resultOfPrevious === null) {
        const matchedParams = matcher.matchUrl(hash);
        return matchedParams === null ? null : handler(matchedParams);
      } else {
        return resultOfPrevious;
      }
    },
    registerRoute<U extends RouteParamsSpec>(matcher: Route<U>, handler: (params: SpecToType<U>) => T): Router<T> {
      return mkNextRouter(this, matcher, handler);
    }
  };
}


/* Helper function for parsing URL's */
/* TODO make more robust. Breaks eg when parameter name is substring of route name */
function findInHash(_hash: string, _key: string): string | null {
  const hash = _hash.toLowerCase();
  const key = _key.toLowerCase();

  const startIndexOfString = hash.indexOf(key);
  if (startIndexOfString === -1) { return null; }

  const rest = hash.substring(startIndexOfString + key.length + 1); // + 1 to account for slash after key

  const strippedRest = rest.indexOf("/") > -1 ? rest.substr(0, rest.indexOf("/")) : rest;
  return strippedRest;
}

function startsWith(big: string, small: string): boolean {
  return big.substr(0, small.length) === small;
}
