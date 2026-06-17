"use client";

import { useState, type FormEvent } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { AuthShell } from "@/components/AuthShell";
import { Field } from "@/components/Field";
import { useAuth } from "@/lib/auth-context";
import { ApiError } from "@/lib/api";

export default function LoginPage() {
  const router = useRouter();
  const { login } = useAuth();
  const [submitting, setSubmitting] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);
  const [fields, setFields] = useState<Record<string, string>>({});

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setSubmitting(true);
    setFormError(null);
    setFields({});

    const data = new FormData(e.currentTarget);
    try {
      await login(String(data.get("email")), String(data.get("password")));
      router.push("/account");
    } catch (err) {
      if (err instanceof ApiError) {
        setFields(err.fields ?? {});
        setFormError(err.fields ? null : err.message);
      } else {
        setFormError("Something went wrong. Is the stack running?");
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <AuthShell
      title="Welcome back"
      subtitle="Sign in to your AURA account."
      highlights={[
        "Tokens are verified by every service, no central session store",
        "Access tokens auto-refresh in the background",
        "Your password is only ever stored as a bcrypt hash",
      ]}
    >
      <form onSubmit={onSubmit} className="space-y-4" noValidate>
        {formError && (
          <div className="rounded-2xl border border-aura-coral/30 bg-aura-coral/10 px-4 py-3 text-sm font-medium text-aura-coral">
            {formError}
          </div>
        )}
        <Field
          label="Email"
          name="email"
          type="email"
          autoComplete="email"
          placeholder="you@example.com"
          required
          error={fields.email}
        />
        <Field
          label="Password"
          name="password"
          type="password"
          autoComplete="current-password"
          placeholder="••••••••"
          required
          error={fields.password}
        />
        <button type="submit" disabled={submitting} className="btn-aura w-full">
          {submitting ? "Signing in…" : "Sign in"}
        </button>
      </form>

      <p className="mt-6 text-sm text-ink-muted">
        New here?{" "}
        <Link href="/register" className="font-semibold text-aura-indigo hover:underline">
          Create an account
        </Link>
      </p>
    </AuthShell>
  );
}
