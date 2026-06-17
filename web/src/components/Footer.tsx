import { Logo } from "./Logo";

const columns = [
  { title: "Shop", links: ["New arrivals", "Audio", "Home", "Wearables"] },
  { title: "Platform", links: ["Architecture", "Services", "API reference", "Status"] },
  { title: "Company", links: ["About", "Careers", "Press", "Contact"] },
];

export function Footer() {
  return (
    <footer className="mt-24 border-t border-ink/8 bg-white/50">
      <div className="container-aura grid gap-10 py-14 md:grid-cols-[1.4fr_repeat(3,1fr)]">
        <div className="max-w-xs">
          <Logo />
          <p className="mt-4 text-sm leading-relaxed text-ink-muted">
            A reference modern commerce platform — microservices, Docker, Go &amp;
            Next.js. Built to learn distributed systems by building one.
          </p>
        </div>
        {columns.map((col) => (
          <div key={col.title}>
            <h4 className="text-xs font-semibold uppercase tracking-wider text-ink-muted">
              {col.title}
            </h4>
            <ul className="mt-4 space-y-2.5">
              {col.links.map((link) => (
                <li key={link}>
                  <span className="cursor-default text-sm text-ink-soft transition hover:text-ink">
                    {link}
                  </span>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
      <div className="border-t border-ink/8">
        <div className="container-aura flex flex-col items-center justify-between gap-2 py-5 text-xs text-ink-muted sm:flex-row">
          <p>© {new Date().getFullYear()} AURA Commerce. A learning project.</p>
          <p>Go · TypeScript · React · Next.js · Docker</p>
        </div>
      </div>
    </footer>
  );
}
