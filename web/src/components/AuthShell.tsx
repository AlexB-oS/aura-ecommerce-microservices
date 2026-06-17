import type { ReactNode } from "react";
import { Logo } from "./Logo";
import { CheckIcon } from "./Icons";

// Shared split-screen layout for the sign-in / sign-up pages: the form on the
// left, a branded "aura" panel with selling points on the right.
export function AuthShell({
  title,
  subtitle,
  children,
  highlights,
}: {
  title: string;
  subtitle: string;
  children: ReactNode;
  highlights: string[];
}) {
  return (
    <div className="container-aura grid min-h-[calc(100vh-4rem)] items-center gap-12 py-12 lg:grid-cols-2">
      {/* Form side */}
      <div className="mx-auto w-full max-w-md">
        <Logo />
        <h1 className="mt-8 font-display text-3xl font-bold tracking-tight">{title}</h1>
        <p className="mt-2 text-ink-muted">{subtitle}</p>
        <div className="mt-8">{children}</div>
      </div>

      {/* Brand side */}
      <div className="relative hidden overflow-hidden rounded-4xl bg-ink p-12 text-white lg:block">
        <div className="pointer-events-none absolute -right-24 -top-24 h-80 w-80 rounded-full bg-aura-violet/40 blur-[100px]" />
        <div className="pointer-events-none absolute -bottom-24 -left-24 h-80 w-80 rounded-full bg-aura-pink/30 blur-[100px]" />
        <div className="pointer-events-none absolute inset-0 bg-dotgrid opacity-30" />
        <div className="relative flex h-full flex-col justify-between">
          <p className="font-display text-2xl font-semibold leading-snug">
            Your account is powered by a real{" "}
            <span className="text-gradient">User microservice</span> — JWT auth,
            bcrypt hashing, the works.
          </p>
          <ul className="mt-10 space-y-4">
            {highlights.map((h) => (
              <li key={h} className="flex items-center gap-3 text-white/80">
                <span className="grid h-6 w-6 shrink-0 place-items-center rounded-full bg-white/10">
                  <CheckIcon width={14} height={14} />
                </span>
                {h}
              </li>
            ))}
          </ul>
        </div>
      </div>
    </div>
  );
}
