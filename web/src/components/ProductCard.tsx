"use client";

import { gradientStyle, type Product } from "@/lib/products";
import { formatPrice } from "@/lib/format";
import { useCart } from "@/lib/cart-context";

export function ProductCard({ product }: { product: Product }) {
  const { add } = useCart();

  return (
    <article className="card group overflow-hidden p-3 transition-all duration-300 hover:-translate-y-1 hover:shadow-card">
      {/* "Photo": a gradient swatch so the repo needs no image assets. */}
      <div
        className="relative aspect-[4/3] w-full overflow-hidden rounded-2xl"
        style={gradientStyle(product)}
      >
        <div className="absolute inset-0 bg-dotgrid opacity-60" />
        {product.badge && (
          <span className="absolute left-3 top-3 rounded-full bg-white/90 px-2.5 py-1 text-[11px] font-bold uppercase tracking-wide text-ink">
            {product.badge}
          </span>
        )}
        <span className="absolute right-3 top-3 rounded-full bg-black/20 px-2.5 py-1 text-[11px] font-medium text-white backdrop-blur">
          {product.category}
        </span>
      </div>

      <div className="px-2 pb-1 pt-4">
        <div className="flex items-start justify-between gap-3">
          <h3 className="font-display text-base font-semibold leading-tight">{product.name}</h3>
          <span className="shrink-0 font-semibold">{formatPrice(product.priceCents)}</span>
        </div>
        <p className="mt-1.5 text-sm leading-relaxed text-ink-muted">{product.tagline}</p>

        <button
          type="button"
          onClick={() => add(product)}
          className="mt-4 w-full rounded-full border border-ink/12 bg-white py-2.5 text-sm font-semibold text-ink transition-all hover:border-transparent hover:bg-aura-gradient hover:text-white hover:shadow-glow"
        >
          Add to cart
        </button>
      </div>
    </article>
  );
}
