import Link from "next/link";

// The AURA wordmark: a gradient "aura" ring glyph next to the name. Kept as a
// component so the brand is consistent across the header, footer and auth pages.
export function Logo({ className = "" }: { className?: string }) {
  return (
    <Link href="/" className={`group inline-flex items-center gap-2.5 ${className}`}>
      <span className="relative grid h-9 w-9 place-items-center">
        <span className="absolute inset-0 rounded-full bg-aura-gradient opacity-90 blur-[2px] transition group-hover:blur-[4px]" />
        <span className="absolute inset-[3px] rounded-full bg-paper" />
        <span className="relative h-2.5 w-2.5 rounded-full bg-aura-gradient" />
      </span>
      <span className="text-lg font-semibold tracking-tight font-display">AURA</span>
    </Link>
  );
}
