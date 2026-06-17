"use client";

import { Suspense, useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useAuth } from "@/lib/auth-context";
import { fetchOrders } from "@/lib/orders-api";
import { formatPrice } from "@/lib/format";
import type { Order } from "@/lib/types";

const statusStyles: Record<Order["status"], string> = {
  paid: "bg-aura-mint/15 text-emerald-600",
  pending_payment: "bg-amber-100 text-amber-700",
  payment_failed: "bg-aura-coral/15 text-aura-coral",
  cancelled: "bg-ink/5 text-ink-muted",
};

const statusLabel: Record<Order["status"], string> = {
  paid: "Paid",
  pending_payment: "Processing payment",
  payment_failed: "Payment failed",
  cancelled: "Cancelled",
};

function OrdersInner() {
  const router = useRouter();
  const params = useSearchParams();
  const { status } = useAuth();
  const placedId = params.get("placed");

  const [orders, setOrders] = useState<Order[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(() => {
    fetchOrders()
      .then(setOrders)
      .catch(() => setError("Could not load your orders."));
  }, []);

  useEffect(() => {
    if (status === "guest") {
      router.replace("/login");
      return;
    }
    if (status === "authenticated") load();
  }, [status, router, load]);

  // Payment is processed asynchronously, so a just-placed order may still be
  // "processing payment". Re-fetch a couple of times to watch it flip to paid.
  useEffect(() => {
    if (!placedId) return;
    const timers = [1500, 3500].map((ms) => setTimeout(load, ms));
    return () => timers.forEach(clearTimeout);
  }, [placedId, load]);

  if (status !== "authenticated") {
    return (
      <div className="flex min-h-[50vh] items-center justify-center">
        <span className="h-8 w-8 animate-spin-slow rounded-full border-2 border-ink/20 border-t-aura-violet" />
      </div>
    );
  }

  return (
    <div className="container-aura max-w-3xl py-16">
      <div className="flex items-center justify-between">
        <h1 className="font-display text-4xl font-bold tracking-tight">Your orders</h1>
        <button onClick={load} className="btn-ghost px-4 py-2 text-sm">
          Refresh
        </button>
      </div>

      {placedId && (
        <div className="mt-6 rounded-2xl border border-aura-mint/30 bg-aura-mint/10 px-5 py-4">
          <p className="font-semibold text-emerald-700">Order placed 🎉</p>
          <p className="mt-1 text-sm text-ink-soft">
            We&apos;re processing your payment now — this page updates automatically.
            Watch the notification feed at{" "}
            <code className="rounded bg-ink/5 px-1.5 py-0.5 text-xs">/api/v1/notifications</code>.
          </p>
        </div>
      )}

      {error && <p className="mt-6 text-sm font-medium text-aura-coral">{error}</p>}

      <div className="mt-8 space-y-3">
        {orders === null && !error ? (
          <p className="text-ink-muted">Loading…</p>
        ) : orders && orders.length === 0 ? (
          <div className="card p-10 text-center">
            <p className="font-medium">No orders yet</p>
            <p className="mt-1 text-sm text-ink-muted">When you check out, your orders show up here.</p>
            <Link href="/products" className="btn-aura mt-5 inline-flex">
              Browse the shop
            </Link>
          </div>
        ) : (
          orders?.map((o) => (
            <div
              key={o.id}
              className={`card flex items-center justify-between p-5 ${
                o.id === placedId ? "ring-2 ring-aura-violet/40" : ""
              }`}
            >
              <div>
                <p className="font-mono text-xs text-ink-muted">#{o.id.slice(0, 8)}</p>
                <p className="mt-1 font-semibold">{formatPrice(o.totalCents)}</p>
                <p className="text-sm text-ink-muted">
                  {new Date(o.createdAt).toLocaleString()}
                </p>
              </div>
              <span
                className={`shrink-0 rounded-full px-3 py-1 text-xs font-semibold ${statusStyles[o.status]}`}
              >
                {statusLabel[o.status]}
              </span>
            </div>
          ))
        )}
      </div>
    </div>
  );
}

export default function OrdersPage() {
  return (
    <Suspense fallback={null}>
      <OrdersInner />
    </Suspense>
  );
}
