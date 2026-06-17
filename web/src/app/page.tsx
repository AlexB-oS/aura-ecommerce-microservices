import Link from "next/link";
import { ProductCard } from "@/components/ProductCard";
import { ArrowRight, BoltIcon, LayersIcon, ShieldIcon, SparkleIcon } from "@/components/Icons";
import { products } from "@/lib/products";

const features = [
  {
    icon: LayersIcon,
    title: "Independent services",
    body: "Each domain — users, catalog, cart, orders — is its own deployable service with its own database.",
  },
  {
    icon: BoltIcon,
    title: "Event-driven",
    body: "Services stay decoupled by publishing events to NATS instead of calling each other directly.",
  },
  {
    icon: ShieldIcon,
    title: "Secure by default",
    body: "Stateless JWT auth, bcrypt password hashing, and a single hardened gateway as the front door.",
  },
  {
    icon: SparkleIcon,
    title: "Built to scale",
    body: "Stateless containers behind NGINX scale horizontally — run three of any service with one flag.",
  },
];

// The six core microservices, with the one that's actually implemented marked
// "Live" — this doubles as a roadmap for the project.
const services = [
  { name: "User", desc: "Registration, auth, profiles", status: "Live" },
  { name: "Product Catalog", desc: "Listings, categories, inventory", status: "Live" },
  { name: "Shopping Cart", desc: "Carts backed by Redis", status: "Live" },
  { name: "Order", desc: "Checkout, order history, status", status: "Live" },
  { name: "Payment", desc: "Simulated charge on order.placed", status: "Live" },
  { name: "Notification", desc: "Email/SMS on events (mocked)", status: "Live" },
];

export default function HomePage() {
  return (
    <>
      {/* ---------------------------------------------------------- Hero --- */}
      <section className="relative overflow-hidden bg-ink text-white">
        {/* aura glows */}
        <div className="pointer-events-none absolute -left-40 -top-40 h-[34rem] w-[34rem] rounded-full bg-aura-violet/40 blur-[120px]" />
        <div className="pointer-events-none absolute -right-32 top-10 h-[28rem] w-[28rem] rounded-full bg-aura-pink/30 blur-[120px]" />
        <div className="pointer-events-none absolute bottom-0 left-1/3 h-[24rem] w-[24rem] rounded-full bg-aura-coral/20 blur-[120px]" />
        <div className="pointer-events-none absolute inset-0 bg-dotgrid opacity-40" />

        <div className="container-aura relative grid gap-12 py-24 md:grid-cols-2 md:items-center md:py-32">
          <div className="animate-fade-up">
            <span className="inline-flex items-center gap-2 rounded-full border border-white/15 bg-white/5 px-3.5 py-1.5 text-xs font-medium text-white/80 backdrop-blur">
              <SparkleIcon width={14} height={14} />
              Microservices · Docker · Go · Next.js
            </span>
            <h1 className="mt-6 font-display text-5xl font-bold leading-[1.05] tracking-tight sm:text-6xl">
              Modern essentials,
              <br />
              <span className="text-gradient">delivered with aura.</span>
            </h1>
            <p className="mt-6 max-w-md text-lg leading-relaxed text-white/70">
              A reference storefront powered by a real distributed system — built
              service by service so you can see how it all fits together.
            </p>
            <div className="mt-9 flex flex-wrap items-center gap-3">
              <Link href="/products" className="btn-aura">
                Browse the shop <ArrowRight width={18} height={18} />
              </Link>
              <Link
                href="/register"
                className="inline-flex items-center justify-center gap-2 rounded-full border border-white/20 bg-white/5 px-6 py-3 text-sm font-semibold text-white backdrop-blur transition hover:bg-white/10"
              >
                Create an account
              </Link>
            </div>
            <p className="mt-8 text-sm text-white/50">
              All six microservices are live — sign in, fill a cart, and check out
              for real (the payment runs as an event-driven saga).
            </p>
          </div>

          {/* Floating product preview cards */}
          <div className="relative hidden h-[26rem] md:block">
            {products.slice(0, 3).map((p, i) => (
              <div
                key={p.id}
                className="absolute w-56 rounded-3xl border border-white/10 bg-white/10 p-3 backdrop-blur-md animate-float"
                style={{
                  top: `${i * 5.5 + 1}rem`,
                  left: `${i * 6 + 1}rem`,
                  animationDelay: `${i * 0.8}s`,
                }}
              >
                <div
                  className="aspect-[5/4] w-full rounded-2xl"
                  style={{
                    backgroundImage: `linear-gradient(135deg, ${p.gradient[0]}, ${p.gradient[1]})`,
                  }}
                />
                <div className="flex items-center justify-between px-1 pt-3 text-white">
                  <span className="text-sm font-semibold">{p.name.split(" ")[0]}</span>
                  <span className="text-sm font-medium text-white/70">
                    ${(p.priceCents / 100).toFixed(0)}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ------------------------------------------------ Featured products --- */}
      <section className="container-aura py-20">
        <div className="flex items-end justify-between gap-4">
          <div>
            <h2 className="font-display text-3xl font-bold tracking-tight">Featured</h2>
            <p className="mt-2 text-ink-muted">A taste of the catalog.</p>
          </div>
          <Link
            href="/products"
            className="hidden items-center gap-1.5 text-sm font-semibold text-aura-indigo transition hover:gap-2.5 sm:inline-flex"
          >
            View all <ArrowRight width={16} height={16} />
          </Link>
        </div>
        <div className="mt-10 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {products.slice(0, 3).map((p) => (
            <ProductCard key={p.id} product={p} />
          ))}
        </div>
      </section>

      {/* --------------------------------------------------------- Features --- */}
      <section id="features" className="container-aura py-12">
        <div className="card overflow-hidden p-8 sm:p-12">
          <h2 className="max-w-2xl font-display text-3xl font-bold tracking-tight">
            A storefront on top of a real platform
          </h2>
          <p className="mt-3 max-w-2xl text-ink-muted">
            Everything you see is served through an API gateway that routes to
            independent microservices. Here is what makes it tick.
          </p>
          <div className="mt-10 grid gap-8 sm:grid-cols-2 lg:grid-cols-4">
            {features.map(({ icon: Icon, title, body }) => (
              <div key={title}>
                <span className="grid h-11 w-11 place-items-center rounded-2xl bg-aura-gradient text-white shadow-glow">
                  <Icon />
                </span>
                <h3 className="mt-4 font-display font-semibold">{title}</h3>
                <p className="mt-1.5 text-sm leading-relaxed text-ink-muted">{body}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* --------------------------------------------------------- Services --- */}
      <section id="services" className="container-aura py-20">
        <div className="mx-auto max-w-2xl text-center">
          <h2 className="font-display text-3xl font-bold tracking-tight">The six core services</h2>
          <p className="mt-3 text-ink-muted">
            Each is built, containerized and scaled independently — and every one
            of them is now live.
          </p>
        </div>
        <div className="mt-12 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {services.map((s) => (
            <div
              key={s.name}
              className="card flex items-center justify-between p-5 transition hover:shadow-card"
            >
              <div>
                <h3 className="font-display font-semibold">{s.name}</h3>
                <p className="mt-0.5 text-sm text-ink-muted">{s.desc}</p>
              </div>
              <span
                className={`shrink-0 rounded-full px-3 py-1 text-xs font-semibold ${
                  s.status === "Live"
                    ? "bg-aura-mint/15 text-emerald-600"
                    : "bg-ink/5 text-ink-muted"
                }`}
              >
                {s.status}
              </span>
            </div>
          ))}
        </div>
      </section>

      {/* ------------------------------------------------------------- CTA --- */}
      <section className="container-aura pb-8">
        <div className="relative overflow-hidden rounded-4xl bg-ink px-8 py-16 text-center text-white sm:px-16">
          <div className="pointer-events-none absolute -left-20 top-0 h-72 w-72 rounded-full bg-aura-violet/40 blur-[100px]" />
          <div className="pointer-events-none absolute -right-20 bottom-0 h-72 w-72 rounded-full bg-aura-coral/30 blur-[100px]" />
          <h2 className="relative font-display text-3xl font-bold tracking-tight sm:text-4xl">
            Ready to look around?
          </h2>
          <p className="relative mx-auto mt-3 max-w-md text-white/70">
            Create an account — it hits the real User microservice through the
            gateway — then explore the shop.
          </p>
          <div className="relative mt-8 flex flex-wrap justify-center gap-3">
            <Link href="/register" className="btn-aura">
              Get started
            </Link>
            <Link
              href="/products"
              className="inline-flex items-center justify-center gap-2 rounded-full border border-white/20 px-6 py-3 text-sm font-semibold text-white transition hover:bg-white/10"
            >
              Browse products
            </Link>
          </div>
        </div>
      </section>
    </>
  );
}
