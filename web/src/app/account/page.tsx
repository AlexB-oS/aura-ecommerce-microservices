"use client";

import { useEffect, useState, type FormEvent } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Field } from "@/components/Field";
import { useAuth } from "@/lib/auth-context";
import { ApiError } from "@/lib/api";

export default function AccountPage() {
  const router = useRouter();
  const { user, status, logout, updateProfile } = useAuth();

  const [name, setName] = useState("");
  const [saving, setSaving] = useState(false);
  const [savedAt, setSavedAt] = useState<number | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Redirect guests to sign-in. This runs once auth state resolves.
  useEffect(() => {
    if (status === "guest") router.replace("/login");
  }, [status, router]);

  // Seed the editable name field once the user loads.
  useEffect(() => {
    if (user) setName(user.name);
  }, [user]);

  if (status !== "authenticated" || !user) {
    return (
      <div className="container-aura flex min-h-[60vh] items-center justify-center">
        <span className="h-8 w-8 animate-spin-slow rounded-full border-2 border-ink/20 border-t-aura-violet" />
      </div>
    );
  }

  async function onSave(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setSaving(true);
    setError(null);
    try {
      await updateProfile(name.trim());
      setSavedAt(Date.now());
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Could not save changes.");
    } finally {
      setSaving(false);
    }
  }

  const memberSince = new Date(user.createdAt).toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });

  return (
    <div className="container-aura max-w-3xl py-16">
      <div className="flex items-center gap-4">
        <span className="grid h-16 w-16 place-items-center rounded-3xl bg-aura-gradient text-2xl font-bold text-white shadow-glow">
          {user.name.charAt(0).toUpperCase()}
        </span>
        <div>
          <h1 className="font-display text-3xl font-bold tracking-tight">{user.name}</h1>
          <p className="text-ink-muted">{user.email}</p>
        </div>
      </div>

      <div className="mt-10 grid gap-6 md:grid-cols-2">
        {/* Editable profile */}
        <form onSubmit={onSave} className="card p-6">
          <h2 className="font-display text-lg font-semibold">Profile</h2>
          <p className="mt-1 text-sm text-ink-muted">
            Updating your name calls{" "}
            <code className="rounded bg-ink/5 px-1.5 py-0.5 text-xs">PATCH /users/me</code>.
          </p>
          <div className="mt-5 space-y-4">
            <Field
              label="Display name"
              name="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              minLength={2}
            />
            <div>
              <label className="label">Email</label>
              <input className="input opacity-60" value={user.email} disabled readOnly />
            </div>
            {error && <p className="text-sm font-medium text-aura-coral">{error}</p>}
            <button
              type="submit"
              disabled={saving || name.trim() === user.name}
              className="btn-aura w-full"
            >
              {saving ? "Saving…" : savedAt ? "Saved ✓" : "Save changes"}
            </button>
          </div>
        </form>

        {/* Account meta */}
        <div className="space-y-6">
          <div className="card p-6">
            <h2 className="font-display text-lg font-semibold">Account</h2>
            <dl className="mt-5 space-y-3 text-sm">
              <div className="flex justify-between">
                <dt className="text-ink-muted">Role</dt>
                <dd className="font-medium capitalize">{user.role}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-ink-muted">Member since</dt>
                <dd className="font-medium">{memberSince}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-ink-muted">User ID</dt>
                <dd className="max-w-[10rem] truncate font-mono text-xs" title={user.id}>
                  {user.id}
                </dd>
              </div>
            </dl>
          </div>

          <Link href="/orders" className="btn-ghost w-full">
            Order history
          </Link>

          <button
            onClick={() => {
              logout();
              router.push("/");
            }}
            className="btn-ghost w-full"
          >
            Sign out
          </button>
        </div>
      </div>
    </div>
  );
}
