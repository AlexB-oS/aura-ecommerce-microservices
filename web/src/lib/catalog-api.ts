// Client for the Product Catalog service (/api/v1/products).
//
// It maps the API's product shape onto the frontend `Product` type the UI
// already renders. The backend stores no presentation data, so the signature
// "aura" gradient for each card is derived deterministically from the product
// slug here — same product, same gradient, every render.

import type { Product } from "./products";

const API_BASE = process.env.NEXT_PUBLIC_API_BASE ?? "";

// The JSON shape returned by the product service.
interface ApiProduct {
  id: string;
  slug: string;
  name: string;
  tagline: string;
  description: string;
  category: string;
  priceCents: number;
  currency: string;
  stock: number;
  badge?: string;
}

// AURA gradient palette; a product's gradient is chosen by hashing its slug.
const palettes: [string, string][] = [
  ["#6D5EFA", "#FF63C3"],
  ["#4B46E5", "#37D9B8"],
  ["#FF7A59", "#FF63C3"],
  ["#FF7A59", "#6D5EFA"],
  ["#37D9B8", "#4B46E5"],
  ["#6D5EFA", "#37D9B8"],
];

function gradientFor(seed: string): [string, string] {
  let h = 0;
  for (let i = 0; i < seed.length; i++) {
    h = (h * 31 + seed.charCodeAt(i)) >>> 0;
  }
  return palettes[h % palettes.length];
}

function mapProduct(p: ApiProduct): Product {
  return {
    id: p.slug,
    name: p.name,
    tagline: p.tagline,
    category: p.category,
    priceCents: p.priceCents,
    gradient: gradientFor(p.slug),
    badge: p.badge || undefined,
  };
}

/** Fetch products from the catalog service. Throws on a non-2xx response so the
 *  caller can fall back to demo data. */
export async function fetchProducts(opts?: {
  category?: string;
  q?: string;
}): Promise<Product[]> {
  const qs = new URLSearchParams();
  if (opts?.category && opts.category !== "All") qs.set("category", opts.category);
  if (opts?.q) qs.set("q", opts.q);

  const url = `${API_BASE}/api/v1/products${qs.toString() ? `?${qs.toString()}` : ""}`;
  const res = await fetch(url, { headers: { "Content-Type": "application/json" } });
  if (!res.ok) throw new Error(`products request failed: ${res.status}`);

  const json = await res.json();
  return (json.data as ApiProduct[]).map(mapProduct);
}
