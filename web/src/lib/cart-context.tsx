"use client";

// A purely client-side demo cart persisted to localStorage.
//
// This stands in for the future Shopping Cart microservice. The UI talks to
// this context exactly like it would talk to a `useCart()` backed by the real
// service, so when you build cart-service the swap is contained to this file:
// replace the localStorage reads/writes with fetches to /api/v1/cart.

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import type { Product } from "./products";

export interface CartItem {
  id: string;
  name: string;
  priceCents: number;
  gradient: [string, string];
  qty: number;
}

interface CartContextValue {
  items: CartItem[];
  count: number;
  totalCents: number;
  isOpen: boolean;
  setOpen: (open: boolean) => void;
  add: (product: Product) => void;
  setQty: (id: string, qty: number) => void;
  remove: (id: string) => void;
  clear: () => void;
}

const CartContext = createContext<CartContextValue | null>(null);
const STORAGE_KEY = "aura.cart";

export function CartProvider({ children }: { children: ReactNode }) {
  const [items, setItems] = useState<CartItem[]>([]);
  const [isOpen, setOpen] = useState(false);

  // Hydrate from localStorage once on mount.
  useEffect(() => {
    try {
      const raw = window.localStorage.getItem(STORAGE_KEY);
      if (raw) setItems(JSON.parse(raw) as CartItem[]);
    } catch {
      /* ignore corrupt storage */
    }
  }, []);

  // Persist on every change.
  useEffect(() => {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(items));
  }, [items]);

  const add = useCallback((product: Product) => {
    setItems((prev) => {
      const existing = prev.find((i) => i.id === product.id);
      if (existing) {
        return prev.map((i) => (i.id === product.id ? { ...i, qty: i.qty + 1 } : i));
      }
      return [
        ...prev,
        {
          id: product.id,
          name: product.name,
          priceCents: product.priceCents,
          gradient: product.gradient,
          qty: 1,
        },
      ];
    });
    setOpen(true);
  }, []);

  const setQty = useCallback((id: string, qty: number) => {
    setItems((prev) =>
      qty <= 0
        ? prev.filter((i) => i.id !== id)
        : prev.map((i) => (i.id === id ? { ...i, qty } : i)),
    );
  }, []);

  const remove = useCallback((id: string) => {
    setItems((prev) => prev.filter((i) => i.id !== id));
  }, []);

  const clear = useCallback(() => setItems([]), []);

  const { count, totalCents } = useMemo(
    () => ({
      count: items.reduce((n, i) => n + i.qty, 0),
      totalCents: items.reduce((n, i) => n + i.qty * i.priceCents, 0),
    }),
    [items],
  );

  return (
    <CartContext.Provider
      value={{ items, count, totalCents, isOpen, setOpen, add, setQty, remove, clear }}
    >
      {children}
    </CartContext.Provider>
  );
}

export function useCart(): CartContextValue {
  const ctx = useContext(CartContext);
  if (!ctx) throw new Error("useCart must be used within a <CartProvider>");
  return ctx;
}
