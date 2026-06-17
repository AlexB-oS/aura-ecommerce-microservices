"use client";

import { useState, type FormEvent } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { AuthShell } from "@/components/AuthShell";
import { Field } from "@/components/Field";
import { useAuth } from "@/lib/auth-context";
import { ApiError } from "@/lib/api";

export default function RegisterPage() {
  const router = useRouter();
  const { register } = useAuth();
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
      await register({
        name: String(data.get("name")),
        email: String(data.get("email")),
        password: String(data.get("password")),
      });
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
      title="Create your account"
      subtitle="It takes a few seconds — and it hits the real User service."
      highlights={[
        "A user.registered event is published to NATS on signup",
        "Validation errors map field-by-field onto this form",
        "You're signed in instantly with a fresh token pair",
      ]}
    >
      <form onSubmit={onSubmit} className="space-y-4" noValidate>
        {formError && (
          <div className="rounded-2xl border border-aura-coral/30 bg-aura-coral/10 px-4 py-3 text-sm font-medium text-aura-coral">
            {formError}
          </div>
        )}
        <Field
          label="Full name"
          name="name"
          type="text"
          autoComplete="name"
          placeholder="Ada Lovelace"
          required
          error={fields.name}
        />
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
          autoComplete="new-password"
          placeholder="At least 8 characters"
          required
          minLength={8}
          error={fields.password}
        />
        <button type="submit" disabled={submitting} className="btn-aura w-full">
          {submitting ? "Creating account…" : "Create account"}
        </button>
      </form>

      <p className="mt-6 text-sm text-ink-muted">
        Already have an account?{" "}
        <Link href="/login" className="font-semibold text-aura-indigo hover:underline">
          Sign in
        </Link>
      </p>
    </AuthShell>
  );
}
