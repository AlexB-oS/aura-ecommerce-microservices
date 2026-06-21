import type { Config } from "tailwindcss";

/**
 * AURA design system.
 * The look: warm paper backgrounds, near-black ink, and a signature "aura"
 * gradient (violet -> pink -> coral) used for glows, highlights and CTAs.
 * Typography pairs Space Grotesk (display) with Inter (body) — see layout.tsx.
 */
const config: Config = {
  content: ["./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        paper: "#F6F4F1",
        // Solid border tints. These exist so component classes can use them
        // inside @apply: Tailwind can't apply an opacity modifier to the
        // DEFAULT shade of a nested color (e.g. `border-ink/12`) within @apply,
        // so we use these flat colors there instead.
        line: {
          DEFAULT: "#E6E3DE",
          strong: "#D4D0CA",
        },
        ink: {
          DEFAULT: "#141318",
          soft: "#3A3A42",
          muted: "#6B6B76",
        },
        aura: {
          violet: "#6D5EFA",
          indigo: "#4B46E5",
          pink: "#FF63C3",
          coral: "#FF7A59",
          mint: "#37D9B8",
        },
      },
      fontFamily: {
        sans: ["var(--font-sans)", "system-ui", "sans-serif"],
        display: ["var(--font-display)", "system-ui", "sans-serif"],
      },
      borderRadius: {
        "4xl": "2rem",
      },
      boxShadow: {
        soft: "0 1px 2px rgba(20,19,24,0.04), 0 8px 30px rgba(20,19,24,0.06)",
        card: "0 2px 4px rgba(20,19,24,0.04), 0 12px 40px rgba(20,19,24,0.08)",
        glow: "0 10px 40px -10px rgba(109,94,250,0.55)",
      },
      backgroundImage: {
        "aura-gradient":
          "linear-gradient(120deg, #6D5EFA 0%, #FF63C3 50%, #FF7A59 100%)",
      },
      keyframes: {
        "fade-up": {
          "0%": { opacity: "0", transform: "translateY(12px)" },
          "100%": { opacity: "1", transform: "translateY(0)" },
        },
        float: {
          "0%, 100%": { transform: "translateY(0)" },
          "50%": { transform: "translateY(-10px)" },
        },
        "spin-slow": {
          to: { transform: "rotate(360deg)" },
        },
      },
      animation: {
        "fade-up": "fade-up 0.6s cubic-bezier(0.22,1,0.36,1) both",
        float: "float 7s ease-in-out infinite",
        "spin-slow": "spin-slow 1s linear infinite",
      },
    },
  },
  plugins: [],
};

export default config;
