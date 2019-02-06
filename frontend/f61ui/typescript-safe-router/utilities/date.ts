export function toYyyymmdd(d: Date) {
  const year = d.getFullYear();
  const month = pad((d.getMonth() + 1).toString(), 2, "0");
  const day = pad(d.getDate().toString(), 2, "0");
  return `${year}-${month}-${day}`;
};
export function fromYyyymmdd (s: string): Date | null {
  const splitted = s.split("-");
  if (splitted.length !== 3) {return null; }
  return new Date(parseInt(splitted[0], 10), parseInt(splitted[1], 10) - 1, parseInt(splitted[2], 10));
};

export function isValidDate(d: Date) {
  return d instanceof Date && !isNaN(d.valueOf());
}

function pad(n: string, width: number, z: string): string {
  z = z || "0";
  n = n + "";
  return n.length >= width ? n : new Array(width - n.length + 1).join(z) + n;
}
