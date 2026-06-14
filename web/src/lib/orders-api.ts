// Client for the cart + order services, used by the checkout flow and the
// orders page. Everything here is authenticated (via api.authed).

import { authed } from "./api";
import type { Order } from "./types";

interface CheckoutLine {
  slug: string;
  qty: number;
}

/**
 * Run a checkout. To keep the demo's local browsing cart and the server-side
 * Cart microservice in sync, we (1) reset the server cart, (2) push the local
 * lines into it, then (3) place the order. Placing the order publishes
 * `order.placed`, which kicks off the stock / payment / notification saga.
 */
export async function checkout(lines: CheckoutLine[]): Promise<Order> {
  // 1. Start from a clean server cart.
  await authed("/cart", { method: "DELETE" });

  // 2. Mirror the local cart into the Cart service (price/name come from the
  //    Product service, so we only send slug + qty).
  for (const line of lines) {
    await authed("/cart/items", {
      method: "POST",
      body: JSON.stringify({ slug: line.slug, qty: line.qty }),
    });
  }

  // 3. Place the order (no body — it reads the server cart).
  return authed<Order>("/orders", { method: "POST" });
}

/** Fetch the signed-in user's order history (newest first). */
export function fetchOrders(): Promise<Order[]> {
  return authed<Order[]>("/orders");
}

/** Fetch a single order with its line items. */
export function fetchOrder(id: string): Promise<Order> {
  return authed<Order>(`/orders/${id}`);
}
