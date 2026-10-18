import { useId, type InputHTMLAttributes } from "react";

type FieldProps = InputHTMLAttributes<HTMLInputElement> & {
  label: string;
};

export function Field({ label, ...inputProps }: FieldProps) {
  const id = useId();

  return (
    <div className="grid gap-1.5 text-left">
      <label htmlFor={id} className="text-sm font-medium">
        {label}
      </label>
      <input
        id={id}
        className="rounded-md border border-border bg-surface px-3 py-2 text-foreground outline-none focus-visible:ring-2 focus-visible:ring-brand"
        {...inputProps}
      />
    </div>
  );
}
