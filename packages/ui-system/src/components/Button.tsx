"use client";
import { forwardRef, type ButtonHTMLAttributes } from "react";
import { clsx } from "clsx";
import { colors, transition, borderRadius, typography } from "@tiki/design-tokens";

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary" | "outline" | "ghost" | "danger";
  size?: "xs" | "sm" | "md" | "lg";
  isLoading?: boolean;
  fullWidth?: boolean;
}

const variants = {
  primary: `bg-[${colors.primary[500]}] text-white hover:bg-[${colors.primary[600]}] focus:ring-[${colors.primary[500]}]`,
  secondary: `bg-[${colors.secondary[500]}] text-white hover:bg-[${colors.secondary[600]}] focus:ring-[${colors.secondary[500]}]`,
  outline: `border-2 border-[${colors.primary[500]}] text-[${colors.primary[500]}] hover:bg-[${colors.primary[500]}] hover:text-white focus:ring-[${colors.primary[500]}]`,
  ghost: `text-[${colors.gray[500]}] hover:bg-[${colors.gray[100]}] focus:ring-[${colors.gray[300]}]`,
  danger: `bg-[${colors.danger[500]}] text-white hover:bg-[${colors.danger[600]}] focus:ring-[${colors.danger[500]}]`,
};

const sizes = {
  xs: "px-2 py-1 text-xs",
  sm: "px-3 py-1.5 text-sm",
  md: "px-5 py-2.5 text-sm",
  lg: "px-8 py-3 text-base",
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant = "primary", size = "md", isLoading, fullWidth, className, children, disabled, ...props }, ref) => {
    const base = `inline-flex items-center justify-center font-medium rounded transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed`;
    return (
      <button ref={ref} className={clsx(base, variants[variant], sizes[size], fullWidth && "w-full", className)} disabled={disabled || isLoading} {...props}>
        {isLoading && <svg className="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24"><circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" /><path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>}
        {children}
      </button>
    );
  }
);
Button.displayName = "Button";
