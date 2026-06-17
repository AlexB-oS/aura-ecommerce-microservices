"use client";

import { useEffect, useMemo, useState } from "react";
import { ProductCard } from "@/components/ProductCard";
import { products as demoProducts, type Product } from "@/lib/products";
import { fetchProducts } from "@/lib/catalog-api";

export default function ProductsPage() {
  // Start with demo data so the page renders instantly, then replace it with
  // live data from the Product service. If the service isn't running, we keep
  // the demo data and say so — the storefront never shows an empty shop.
  const [items, setItems] = useState<Product[]>(demoProducts);
  const [source, setSource] = useState<"live" | "demo" | "loading">("loading");
  const [active, setActive] = useState("All");

  useEffect(() => {
    let cancelled = false;
    fetchProducts()
      .then((live) => {
        if (cancelled) return;
        if (live.length > 0) setItems(live);
        setSource("live");
      })
      .catch(() => {
        if (!cancelled) setSource("demo");
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const categories = useMemo(
    () => ["All", ...Array.from(new Set(items.map((p) => p.category)))],
    [items],
  );
  const visible = active === "All" ? items : items.filter((p) => p.category === active);

  return (
    <div className="container-aura py-16">
      <header className="max-w-2xl">
        <h1 className="font-display text-4xl font-bold tracking-tight">The shop</h1>
        <p className="mt-3 text-ink-muted">
          {source === "live" ? (
            <>Live catalog, served by the Product microservice at <code className="rounded bg-ink/5 px-1.5 py-0.5 text-sm">/api/v1/products</code>.</>
          ) : source === "demo" ? (
            <>Showing demo data — start the stack (<code className="rounded bg-ink/5 px-1.5 py-0.5 text-sm">docker compose up</code>) to load the live catalog.</>
          ) : (
            <>Loading the catalog…</>
          )}
        </p>
      </header>

      {/* Category filter */}
      <div className="mt-8 flex flex-wrap gap-2">
        {categories.map((cat) => (
          <button
            key={cat}
            onClick={() => setActive(cat)}
            className={`rounded-full px-4 py-2 text-sm font-medium transition ${
              active === cat
                ? "bg-ink text-white"
                : "border border-ink/12 bg-white/60 text-ink-soft hover:bg-white"
            }`}
          >
            {cat}
          </button>
        ))}
      </div>

      <div className="mt-10 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {visible.map((p) => (
          <ProductCard key={p.id} product={p} />
        ))}
      </div>
    </div>
  );
}
