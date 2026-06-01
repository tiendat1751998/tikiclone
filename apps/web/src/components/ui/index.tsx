"use client";

import { forwardRef, type ButtonHTMLAttributes, type InputHTMLAttributes } from "react";
import { clsx } from "clsx";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary" | "outline" | "ghost" | "danger";
  size?: "sm" | "md" | "lg";
  isLoading?: boolean;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
}

const buttonStyles = {
  base: "inline-flex items-center justify-center font-medium rounded-lg transition-all duration-150 focus:outline-none focus:ring-2 focus:ring-offset-1 disabled:opacity-50 disabled:cursor-not-allowed",
  variant: {
    primary: "bg-tiki-blue text-white hover:bg-tiki-blue-dark focus:ring-tiki-blue",
    secondary: "bg-gray-100 text-tiki-text-secondary hover:bg-gray-200 focus:ring-gray-300",
    outline: "border border-tiki-blue text-tiki-blue hover:bg-blue-50 focus:ring-tiki-blue",
    ghost: "text-tiki-text-secondary hover:bg-gray-100 focus:ring-gray-300",
    danger: "bg-tiki-red text-white hover:bg-tiki-red-dark focus:ring-tiki-red",
  },
  size: {
    sm: "px-3 py-1.5 text-xs gap-1",
    md: "px-4 py-2 text-sm gap-1.5",
    lg: "px-6 py-3 text-base gap-2",
  },
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant = "primary", size = "md", isLoading, leftIcon, rightIcon, className, children, disabled, ...props }, ref) => {
    return (
      <button
        ref={ref}
        className={clsx(buttonStyles.base, buttonStyles.variant[variant], buttonStyles.size[size], className)}
        disabled={disabled || isLoading}
        {...props}
      >
        {isLoading ? (
          <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
        ) : leftIcon}
        {children}
        {!isLoading && rightIcon}
      </button>
    );
  }
);
Button.displayName = "Button";

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  helperText?: string;
  leftAddon?: React.ReactNode;
  rightAddon?: React.ReactNode;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ label, error, helperText, leftAddon, rightAddon, className, id, ...props }, ref) => {
    const inputId = id || label?.toLowerCase().replace(/\s+/g, "-");
    return (
      <div className="w-full">
        {label && (
          <label htmlFor={inputId} className="block text-sm font-medium text-tiki-text mb-1">
            {label}
          </label>
        )}
        <div className="relative">
          {leftAddon && (
            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-tiki-text-secondary">
              {leftAddon}
            </div>
          )}
          <input
            ref={ref}
            id={inputId}
            className={clsx(
              "block w-full rounded-lg border border-gray-300 bg-white text-sm text-tiki-text",
              "placeholder:text-gray-400",
              "focus:border-tiki-blue focus:ring-1 focus:ring-tiki-blue focus:outline-none",
              "disabled:bg-gray-50 disabled:text-gray-500 disabled:cursor-not-allowed",
              leftAddon ? "pl-10" : "pl-3",
              rightAddon ? "pr-10" : "pr-3",
              "py-2",
              error && "border-tiki-red focus:border-tiki-red focus:ring-tiki-red",
              className
            )}
            {...props}
          />
          {rightAddon && (
            <div className="absolute inset-y-0 right-0 pr-3 flex items-center text-tiki-text-secondary">
              {rightAddon}
            </div>
          )}
        </div>
        {error && <p className="mt-1 text-xs text-tiki-red">{error}</p>}
        {helperText && !error && <p className="mt-1 text-xs text-tiki-text-secondary">{helperText}</p>}
      </div>
    );
  }
);
Input.displayName = "Input";

interface BadgeProps {
  variant?: "default" | "success" | "warning" | "danger" | "info" | "outline";
  size?: "sm" | "md";
  children: React.ReactNode;
  className?: string;
}

const badgeStyles = {
  variant: {
    default: "bg-gray-100 text-tiki-text-secondary",
    success: "bg-green-100 text-green-700",
    warning: "bg-yellow-100 text-yellow-700",
    danger: "bg-red-100 text-tiki-red",
    info: "bg-blue-100 text-tiki-blue",
    outline: "border border-tiki-border text-tiki-text-secondary",
  },
  size: {
    sm: "px-1.5 py-0.5 text-[10px]",
    md: "px-2 py-0.5 text-xs",
  },
};

export function Badge({ variant = "default", size = "md", children, className }: BadgeProps) {
  return (
    <span className={clsx("inline-flex items-center font-medium rounded", badgeStyles.variant[variant], badgeStyles.size[size], className)}>
      {children}
    </span>
  );
}

interface CardProps {
  children: React.ReactNode;
  className?: string;
  padding?: "none" | "sm" | "md" | "lg";
  hover?: boolean;
}

const cardPadding = {
  none: "",
  sm: "p-3",
  md: "p-4",
  lg: "p-6",
};

export function Card({ children, className, padding = "md", hover = false }: CardProps) {
  return (
    <div className={clsx("bg-white rounded-lg border border-tiki-border", cardPadding[padding], hover && "hover:shadow-md transition-shadow", className)}>
      {children}
    </div>
  );
}

export function Skeleton({ className }: { className?: string }) {
  return <div className={clsx("animate-pulse bg-gray-200 rounded", className)} />;
}

export function ProductCardSkeleton() {
  return (
    <div className="bg-white rounded-lg border border-tiki-border overflow-hidden">
      <Skeleton className="aspect-square w-full" />
      <div className="p-3 space-y-2">
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-2/3" />
        <Skeleton className="h-5 w-1/2" />
      </div>
    </div>
  );
}

export function ProductGridSkeleton({ count = 8 }: { count?: number }) {
  return (
    <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-3">
      {Array.from({ length: count }).map((_, i) => (
        <ProductCardSkeleton key={i} />
      ))}
    </div>
  );
}

interface StarRatingProps {
  rating: number;
  maxStars?: number;
  size?: "sm" | "md" | "lg";
  showValue?: boolean;
  count?: number;
}

const starSize = { sm: "text-xs", md: "text-sm", lg: "text-base" };

export function StarRating({ rating, maxStars = 5, size = "sm", showValue = false, count }: StarRatingProps) {
  return (
    <div className="inline-flex items-center gap-1">
      <div className="flex">
        {Array.from({ length: maxStars }).map((_, i) => (
          <span key={i} className={clsx(starSize[size], i < Math.round(rating) ? "text-yellow-400" : "text-gray-300")}>★</span>
        ))}
      </div>
      {showValue && rating > 0 && <span className="text-xs text-tiki-text-secondary ml-1">{rating.toFixed(1)}</span>}
      {count !== undefined && <span className="text-xs text-tiki-text-secondary ml-1">({count})</span>}
    </div>
  );
}

interface PriceProps {
  amount: number;
  originalAmount?: number | null;
  discountPercent?: number | null;
  size?: "sm" | "md" | "lg";
  className?: string;
}

const priceSize = { sm: "text-sm", md: "text-base font-bold", lg: "text-xl font-bold" };

export function Price({ amount, originalAmount, discountPercent, size = "md", className }: PriceProps) {
  return (
    <div className={clsx("inline-flex items-center gap-2", className)}>
      <span className={clsx(priceSize[size], "text-tiki-red")}>
        {amount?.toLocaleString("vi-VN")} ₫
      </span>
      {originalAmount && originalAmount > amount && (
        <span className="text-xs text-tiki-text-secondary line-through">
          {originalAmount.toLocaleString("vi-VN")} ₫
        </span>
      )}
      {discountPercent && discountPercent > 0 && (
        <span className="text-[10px] font-bold text-tiki-red bg-red-50 px-1 rounded">
          -{discountPercent}%
        </span>
      )}
    </div>
  );
}

interface EmptyStateProps {
  icon?: React.ReactNode;
  title: string;
  description?: string;
  action?: React.ReactNode;
}

export function EmptyState({ icon, title, description, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-12 px-4 text-center">
      {icon && <div className="text-4xl mb-4 text-gray-300">{icon}</div>}
      <h3 className="text-base font-medium text-tiki-text mb-1">{title}</h3>
      {description && <p className="text-sm text-tiki-text-secondary mb-4 max-w-sm">{description}</p>}
      {action}
    </div>
  );
}

import { createContext, useContext, useState } from "react";

interface TabsContextValue { activeTab: string; setActiveTab: (tab: string) => void; }

const TabsContext = createContext<TabsContextValue | null>(null);

interface TabsProps { defaultValue: string; children: React.ReactNode; className?: string; }

export function Tabs({ defaultValue, children, className }: TabsProps) {
  const [activeTab, setActiveTab] = useState(defaultValue);
  return (
    <TabsContext.Provider value={{ activeTab, setActiveTab }}>
      <div className={className}>{children}</div>
    </TabsContext.Provider>
  );
}

export function TabList({ children, className }: { children: React.ReactNode; className?: string }) {
  return <div className={clsx("flex border-b border-tiki-border", className)} role="tablist">{children}</div>;
}

export function Tab({ value, children, className }: { value: string; children: React.ReactNode; className?: string }) {
  const ctx = useContext(TabsContext);
  if (!ctx) throw new Error("Tab must be used within Tabs");
  const isActive = ctx.activeTab === value;
  return (
    <button
      role="tab"
      aria-selected={isActive}
      onClick={() => ctx.setActiveTab(value)}
      className={clsx(
        "px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors",
        isActive ? "border-tiki-blue text-tiki-blue" : "border-transparent text-tiki-text-secondary hover:text-tiki-text",
        className
      )}
    >
      {children}
    </button>
  );
}

export function TabPanel({ value, children }: { value: string; children: React.ReactNode }) {
  const ctx = useContext(TabsContext);
  if (!ctx) throw new Error("TabPanel must be used within Tabs");
  if (ctx.activeTab !== value) return null;
  return <div className="py-4">{children}</div>;
}

// Re-exports
export { default as Pagination } from "./Pagination";
