"use client";

import Link from "next/link";
import { Logo } from "./Logo";
import { CartIcon, UserIcon } from "./Icons";
import { useAuth } from "@/lib/auth-context";
import { useCart } from "@/lib/cart-context";

const navLinks = [
  { href: "/products", label: "Shop" },
  { href: "/#features", label: "Platform" },
  { href: "/#services", label: "Services" },
];

export function Header() {
  const { user, status, logout } = useAuth();
  const { count, setOpen } = useCart();

  return (
    <header className="sticky top-0 z-40 border-b border-ink/8 bg-paper/80 backdrop-blur-md">
      <div className="container-aura flex h-16 items-center justify-between gap-4">
        <Logo />

        <nav className="hidden items-center gap-8 md:flex">
          {navLinks.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              className="text-sm font-medium text-ink-soft transition hover:text-ink"
            >
              {link.label}
            </Link>
          ))}
        </nav>

        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => setOpen(true)}
            className="relative grid h-10 w-10 place-items-center rounded-full text-ink-soft transition hover:bg-white hover:text-ink"
            aria-label="Open cart"
          >
            <CartIcon />
            {count > 0 && (
              <span className="absolute -right-0.5 -top-0.5 grid h-5 min-w-5 place-items-center rounded-full bg-aura-gradient px-1 text-[11px] font-bold text-white">
                {count}
              </span>
            )}
          </button>

          {status === "authenticated" && user ? (
            <div className="flex items-center gap-2">
              <Link
                href="/account"
                className="flex items-center gap-2 rounded-full border border-ink/10 bg-white/70 py-1.5 pl-1.5 pr-3.5 text-sm font-medium transition hover:bg-white"
              >
                <span className="grid h-7 w-7 place-items-center rounded-full bg-aura-gradient text-white">
                  <UserIcon width={15} height={15} />
                </span>
                <span className="hidden sm:inline">{user.name.split(" ")[0]}</span>
              </Link>
              <button
                type="button"
                onClick={logout}
                className="hidden text-sm font-medium text-ink-muted transition hover:text-ink sm:inline"
              >
                Sign out
              </button>
            </div>
          ) : status === "loading" ? (
            <span className="h-10 w-24 animate-pulse rounded-full bg-ink/5" />
          ) : (
            <>
              <Link href="/login" className="btn-ghost hidden px-5 py-2 sm:inline-flex">
                Sign in
              </Link>
              <Link href="/register" className="btn-aura px-5 py-2">
                Get started
              </Link>
            </>
          )}
        </div>
      </div>
    </header>
  );
}
