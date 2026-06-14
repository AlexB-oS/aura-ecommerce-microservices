// Money is stored as integer cents everywhere (never floats) to avoid rounding
// errors. This helper renders cents as a localized currency string.
export function formatPrice(cents: number, currency = "USD"): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency,
  }).format(cents / 100);
}
