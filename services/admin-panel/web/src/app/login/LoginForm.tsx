'use client';

import { useState, useTransition, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@shopee/shared-auth';
import { Button, Input } from '@shopee/ui-system';
import { LoginSchema, type LoginFormData } from '@/lib/validations';
import { z } from 'zod';

export function LoginForm() {
  const router = useRouter();
  const [isPending, startTransition] = useTransition();
  const [error, setError] = useState<string | null>(null);
  const [formData, setFormData] = useState<LoginFormData>({
    email: '',
    password: '',
  });
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  const login = useAuthStore((state) => state.login);

  const handleChange = useCallback((field: keyof LoginFormData, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    if (fieldErrors[field]) {
      setFieldErrors((prev) => {
        const updated = { ...prev };
        delete updated[field];
        return updated;
      });
    }
    setError(null);
  }, [fieldErrors]);

  const handleSubmit = useCallback(async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    const result = LoginSchema.safeParse(formData);
    if (!result.success) {
      const errors: Record<string, string> = {};
      result.error.issues.forEach((issue) => {
        const field = issue.path[0] as string;
        if (!errors[field]) {
          errors[field] = issue.message;
        }
      });
      setFieldErrors(errors);
      return;
    }

    startTransition(async () => {
      try {
        const response = await fetch('/api/auth/login', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'X-Correlation-ID': `login-${Date.now()}`,
          },
          body: JSON.stringify({
            email: formData.email,
            password: formData.password,
          }),
        });

        if (!response.ok) {
          const errorData = await response.json().catch(() => ({}));
          throw new Error(errorData.message || 'Authentication failed');
        }

        const data = await response.json();
        const authData = data.data || data;

        if (authData.user && authData.access_token) {
          login(
            {
              id: authData.user.id,
              email: authData.user.email,
              username: authData.user.username || authData.user.email.split('@')[0],
              display_name: authData.user.display_name || authData.user.username || 'Admin',
              phone: authData.user.phone || '',
              avatar_url: authData.user.avatar_url || '',
              status: authData.user.status || 'active',
              created_at: authData.user.created_at || new Date().toISOString(),
              role: authData.user.role || 'viewer',
            },
            {
              access_token: authData.access_token,
              refresh_token: authData.refresh_token,
              expires_in: authData.expires_in || 3600,
              token_type: authData.token_type || 'Bearer',
            }
          );

          router.push('/admin');
          router.refresh();
        } else {
          throw new Error('Invalid response from server');
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Authentication failed');
      }
    });
  }, [formData, login, router]);

  return (
    <form onSubmit={handleSubmit} className="space-y-5">
      {error && (
        <div className="p-3 rounded-lg bg-danger-50 dark:bg-danger-900/20 border border-danger-200 dark:border-danger-800">
          <p className="text-sm text-danger-600 dark:text-danger-400">{error}</p>
        </div>
      )}

      <div>
        <label htmlFor="email" className="block text-sm font-medium text-foreground mb-1.5">
          Email Address
        </label>
        <input
          id="email"
          type="email"
          autoComplete="email"
          value={formData.email}
          onChange={(e) => handleChange('email', e.target.value)}
          className={`w-full px-3 py-2.5 rounded-lg border bg-card text-sm transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent ${
            fieldErrors.email
              ? 'border-danger-500'
              : 'border-border hover:border-primary-300'
          }`}
          placeholder="admin@tiki.vn"
          disabled={isPending}
        />
        {fieldErrors.email && (
          <p className="mt-1 text-xs text-danger-500">{fieldErrors.email}</p>
        )}
      </div>

      <div>
        <label htmlFor="password" className="block text-sm font-medium text-foreground mb-1.5">
          Password
        </label>
        <input
          id="password"
          type="password"
          autoComplete="current-password"
          value={formData.password}
          onChange={(e) => handleChange('password', e.target.value)}
          className={`w-full px-3 py-2.5 rounded-lg border bg-card text-sm transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent ${
            fieldErrors.password
              ? 'border-danger-500'
              : 'border-border hover:border-primary-300'
          }`}
          placeholder="Enter your password"
          disabled={isPending}
        />
        {fieldErrors.password && (
          <p className="mt-1 text-xs text-danger-500">{fieldErrors.password}</p>
        )}
      </div>

      <button
        type="submit"
        disabled={isPending}
        className="w-full py-2.5 px-4 rounded-lg bg-primary-500 text-white font-medium text-sm hover:bg-primary-600 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        {isPending ? (
          <span className="inline-flex items-center gap-2">
            <svg className="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
            </svg>
            Signing in...
          </span>
        ) : (
          'Sign In'
        )}
      </button>
    </form>
  );
}
