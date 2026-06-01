"use client";
import { forwardRef, type InputHTMLAttributes } from "react";
import { clsx } from "clsx";
import { colors } from "@tiki/design-tokens";

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  helperText?: string;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ label, error, helperText, className, id, ...props }, ref) => {
    const inputId = id || label?.toLowerCase().replace(/\s+/g, "-");
    return (
      <div className="w-full">
        {label && <label htmlFor={inputId} className="block text-sm font-medium text-[${colors.text.primary}] mb-1">{label}</label>}
        <input ref={ref} id={inputId} className={clsx(
          "w-full px-3 py-2.5 border rounded text-sm transition-colors",
          "focus:outline-none focus:ring-2 focus:ring-[${colors.primary[500]}] focus:border-transparent",
          error ? "border-[${colors.danger[500]}] focus:ring-[${colors.danger[500]}]" : "border-[${colors.border.default}] hover:border-[${colors.border.hover}]",
          className
        )} {...props} />
        {error && <p className="mt-1 text-xs text-[${colors.danger[500]}]">{error}</p>}
        {helperText && !error && <p className="mt-1 text-xs text-[${colors.text.secondary}]">{helperText}</p>}
      </div>
    );
  }
);
Input.displayName = "Input";
