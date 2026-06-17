// A labelled form input that surfaces the per-field validation message the user
// service returns (httpx error `fields` map -> red helper text under the input).
import type { InputHTMLAttributes } from "react";

interface FieldProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string;
  name: string;
  error?: string;
}

export function Field({ label, name, error, ...rest }: FieldProps) {
  return (
    <div>
      <label htmlFor={name} className="label">
        {label}
      </label>
      <input
        id={name}
        name={name}
        className={`input ${error ? "border-aura-coral/70 focus:border-aura-coral focus:ring-aura-coral/15" : ""}`}
        aria-invalid={error ? true : undefined}
        {...rest}
      />
      {error && <p className="mt-1.5 text-xs font-medium text-aura-coral">{error}</p>}
    </div>
  );
}
