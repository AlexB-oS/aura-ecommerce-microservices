// Demo catalog data.
//
// This is static placeholder data so the storefront looks real before the
// Product Catalog Service exists. When you build that service, replace these
// imports with a fetch to /api/v1/products — the components already render from
// this shape, so the swap is mechanical.
//
// Products use CSS gradients (not image files) as their artwork, so the repo
// stays asset-free and the design feels cohesive with the AURA palette.

import type { CSSProperties } from "react";

export interface Product {
  id: string;
  name: string;
  tagline: string;
  category: string;
  priceCents: number;
  /** two-stop gradient used as the product "photo" */
  gradient: [string, string];
  badge?: string;
}

export const products: Product[] = [
  {
    id: "halo-ambient-lamp",
    name: "Halo Ambient Lamp",
    tagline: "Warm, dimmable glow with a featherlight aluminium body.",
    category: "Home",
    priceCents: 12900,
    gradient: ["#6D5EFA", "#FF63C3"],
    badge: "New",
  },
  {
    id: "drift-headphones",
    name: "Drift Wireless Headphones",
    tagline: "40h playback, adaptive noise cancelling, zero fatigue.",
    category: "Audio",
    priceCents: 21900,
    gradient: ["#4B46E5", "#37D9B8"],
    badge: "Best seller",
  },
  {
    id: "nimbus-daypack",
    name: "Nimbus Daypack",
    tagline: "Weatherproof 22L pack that swallows a 16\" laptop whole.",
    category: "Travel",
    priceCents: 14500,
    gradient: ["#FF7A59", "#FF63C3"],
  },
  {
    id: "ember-temperature-mug",
    name: "Ember Temperature Mug",
    tagline: "Keeps coffee at the exact degree you like, all morning.",
    category: "Home",
    priceCents: 9900,
    gradient: ["#FF7A59", "#6D5EFA"],
  },
  {
    id: "pulse-smart-watch",
    name: "Pulse Smart Watch",
    tagline: "Health tracking and a 9-day battery in 32g of titanium.",
    category: "Wearables",
    priceCents: 27900,
    gradient: ["#37D9B8", "#4B46E5"],
    badge: "New",
  },
  {
    id: "lumen-keyboard",
    name: "Lumen Mechanical Keyboard",
    tagline: "Hot-swappable switches and per-key aura backlighting.",
    category: "Desk",
    priceCents: 16900,
    gradient: ["#6D5EFA", "#37D9B8"],
  },
];

export function gradientStyle(p: Product): CSSProperties {
  return {
    backgroundImage: `linear-gradient(135deg, ${p.gradient[0]}, ${p.gradient[1]})`,
  };
}
