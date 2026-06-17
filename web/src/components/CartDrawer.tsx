"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useCart } from "@/lib/cart-context";
import { useAuth } from "@/lib/auth-context";
import { formatPrice } from "@/lib/format";
import { checkout } from "@/lib/orders-api";
import { ApiError } from "@/lib/api";
import { CloseIcon } from "./Icons";

// Slide-over cart. Browsing uses the local CartProvider; checkout pushes the
// cart into the real Cart service and places an order via the Order service,
// which triggers the stock/payment/notification saga.
export function CartDrawer() {
  const { items, isOpen, setOpen, setQty, remove, totalCents, count, clear } = useCart();
  const { status } = useAuth();
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function onCheckout() {
    setError(null);
    // Checkout requires an account (the cart + orders are user-scoped).
    if (status !== "authenticated") {
      setOpen(false);
      router.push("/login");
      return;
    }
    setBusy(true);
    try {
      const order = await checkout(items.map((i) => ({ slug: i.id, qty: i.qty })));
      clear();
      setOpen(false);
      router.push(`/orders?placed=${order.id}`);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Checkout failed. Is the stack running?");
    } finally {
      setBusy(false);
    }
  }

  return (
    <>
      {/* Backdrop */}
      <div
        className={`fixed inset-0 z-50 bg-ink/30 backdrop-blur-sm transition-opacity duration-300 ${
          isOpen ? "opacity-100" : "pointer-events-none opacity-0"
        }`}
        onClick={() => setOpen(false)}
        aria-hidden
      />

      {/* Panel */}
      <aside
        className={`fixed right-0 top-0 z-50 flex h-full w-full max-w-md flex-col bg-paper shadow-2xl transition-transform duration-300 ${
          isOpen ? "translate-x-0" : "translate-x-full"
        }`}
        role="dialog"
        aria-label="Shopping cart"
      >
        <div className="flex items-center justify-between border-b border-ink/8 px-6 py-5">
          <h2 className="font-display text-lg font-semibold">
            Your cart {count > 0 && <span className="text-ink-muted">({count})</span>}
          </h2>
          <button
            onClick={() => setOpen(false)}
            className="grid h-9 w-9 place-items-center rounded-full text-ink-soft transition hover:bg-white hover:text-ink"
            aria-label="Close cart"
          >
            <CloseIcon />
          </button>
        </div>

        <div className="flex-1 overflow-y-auto px-6 py-4">
          {items.length === 0 ? (
            <div className="flex h-full flex-col items-center justify-center text-center">
              <div className="h-16 w-16 rounded-full bg-aura-gradient opacity-20" />
              <p className="mt-4 font-medium">Your cart is empty</p>
              <p className="mt-1 text-sm text-ink-muted">Add a few things to see them here.</p>
            </div>
          ) : (
            <ul className="space-y-4">
              {items.map((item) => (
                <li key={item.id} className="flex gap-4">
                  <div
                    className="h-16 w-16 shrink-0 rounded-xl"
                    style={{
                      backgroundImage: `linear-gradient(135deg, ${item.gradient[0]}, ${item.gradient[1]})`,
                    }}
                  />
                  <div className="flex flex-1 flex-col">
                    <div className="flex justify-between gap-2">
                      <span className="text-sm font-semibold">{item.name}</span>
                      <span className="text-sm font-semibold">
                        {formatPrice(item.priceCents * item.qty)}
                      </span>
                    </div>
                    <div className="mt-auto flex items-center justify-between">
                      <div className="flex items-center rounded-full border border-ink/12">
                        <button
                          onClick={() => setQty(item.id, item.qty - 1)}
                          className="grid h-7 w-7 place-items-center text-ink-soft hover:text-ink"
                          aria-label="Decrease quantity"
                        >
                          –
                        </button>
                        <span className="w-6 text-center text-sm font-medium">{item.qty}</span>
                        <button
                          onClick={() => setQty(item.id, item.qty + 1)}
                          className="grid h-7 w-7 place-items-center text-ink-soft hover:text-ink"
                          aria-label="Increase quantity"
                        >
                          +
                        </button>
                      </div>
                      <button
                        onClick={() => remove(item.id)}
                        className="text-xs font-medium text-ink-muted hover:text-aura-coral"
                      >
                        Remove
                      </button>
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>

        <div className="border-t border-ink/8 px-6 py-5">
          <div className="mb-1 flex items-center justify-between">
            <span className="text-sm text-ink-muted">Subtotal</span>
            <span className="text-lg font-semibold">{formatPrice(totalCents)}</span>
          </div>
          <p className="mb-4 text-xs text-ink-muted">
            {status === "authenticated"
              ? "Checkout places a real order and runs the payment saga."
              : "You'll be asked to sign in to check out."}
          </p>
          {error && (
            <p className="mb-3 rounded-xl border border-aura-coral/30 bg-aura-coral/10 px-3 py-2 text-xs font-medium text-aura-coral">
              {error}
            </p>
          )}
          <button
            disabled={items.length === 0 || busy}
            onClick={onCheckout}
            className="btn-aura w-full"
          >
            {busy ? "Placing order…" : "Checkout"}
          </button>
        </div>
      </aside>
    </>
  );
}
