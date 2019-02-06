export function safeParseInt(v: string): number | null {
  let parsed = parseInt(v, 10);
  return isInteger(parsed) && !isNaN(parsed) ? parsed : null;
}

export function safeParseNum(v: string): number | null {
  let parsed = parseFloat(v);
  return isNumber(parsed) && !isNaN(parsed) ? parsed : null;
}

function isInteger(n: any): boolean {
  return n === parseInt(n, 10);
}

function isNumber(obj: any): boolean {
  return !isNaN(parseFloat(obj));
}
