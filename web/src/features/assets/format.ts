export function dateOnly(value?: string | null) {
  if (!value) return '';
  return value.slice(0, 10);
}

export function money(value?: number | null) {
  if (value === undefined || value === null) return '';
  return new Intl.NumberFormat(undefined, { style: 'currency', currency: 'USD' }).format(value);
}

export function bytes(value: number) {
  if (value < 1024) return `${value} B`;
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`;
  return `${(value / (1024 * 1024)).toFixed(1)} MB`;
}
