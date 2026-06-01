'use client';

import { useState, useTransition, useCallback } from 'react';
import { cn } from '@shopee/ui-system';

interface SettingsSection {
  id: string;
  title: string;
  description: string;
}

const SETTINGS_SECTIONS: SettingsSection[] = [
  { id: 'general', title: 'General', description: 'Basic store information and preferences' },
  { id: 'appearance', title: 'Appearance', description: 'Customize the look and feel of your admin panel' },
  { id: 'notifications', title: 'Notifications', description: 'Configure email and in-app notifications' },
  { id: 'security', title: 'Security', description: 'Two-factor authentication and session settings' },
  { id: 'api', title: 'API & Integrations', description: 'Manage API keys and third-party integrations' },
  { id: 'billing', title: 'Billing', description: 'Subscription and payment information' },
];

export default function SettingsPage() {
  const [activeSection, setActiveSection] = useState('general');

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-foreground">Settings</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Manage your admin panel configuration and preferences
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        <div className="lg:col-span-1">
          <nav className="space-y-1">
            {SETTINGS_SECTIONS.map((section) => (
              <button
                key={section.id}
                onClick={() => setActiveSection(section.id)}
                className={cn(
                  'w-full text-left px-4 py-3 rounded-lg transition-colors',
                  activeSection === section.id
                    ? 'bg-primary-500/10 text-primary-600 dark:text-primary-400'
                    : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                )}
              >
                <p className="text-sm font-medium">{section.title}</p>
                <p className="text-xs mt-0.5 opacity-70">{section.description}</p>
              </button>
            ))}
          </nav>
        </div>

        <div className="lg:col-span-3">
          {activeSection === 'general' && <GeneralSettings />}
          {activeSection === 'appearance' && <AppearanceSettings />}
          {activeSection === 'notifications' && <NotificationSettings />}
          {activeSection === 'security' && <SecuritySettings />}
          {activeSection === 'api' && <ApiSettings />}
          {activeSection === 'billing' && <BillingSettings />}
        </div>
      </div>
    </div>
  );
}

function GeneralSettings() {
  const [isPending, startTransition] = useTransition();

  return (
    <div className="space-y-6">
      <div className="rounded-xl border border-border bg-card p-6">
        <h3 className="text-lg font-semibold text-foreground mb-4">Store Information</h3>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Store Name</label>
            <input
              type="text"
              defaultValue="Tiki Clone"
              className="w-full px-3 py-2.5 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
              disabled={isPending}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Store URL</label>
            <input
              type="url"
              defaultValue="https://tiki.vn"
              className="w-full px-3 py-2.5 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
              disabled={isPending}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Contact Email</label>
            <input
              type="email"
              defaultValue="admin@tiki.vn"
              className="w-full px-3 py-2.5 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
              disabled={isPending}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Currency</label>
            <select className="w-full px-3 py-2.5 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500" disabled={isPending}>
              <option value="VND">Vietnamese Dong (₫)</option>
              <option value="USD">US Dollar ($)</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Timezone</label>
            <select className="w-full px-3 py-2.5 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500" disabled={isPending}>
              <option value="Asia/Ho_Chi_Minh">Asia/Ho Chi Minh (UTC+7)</option>
              <option value="Asia/Hanoi">Asia/Hanoi (UTC+7)</option>
            </select>
          </div>
        </div>
      </div>

      <div className="rounded-xl border border-border bg-card p-6">
        <h3 className="text-lg font-semibold text-foreground mb-4">Order Settings</h3>
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Order Prefix</label>
            <input
              type="text"
              defaultValue="TK"
              className="w-full px-3 py-2.5 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
              disabled={isPending}
            />
            <p className="mt-1 text-xs text-muted-foreground">Prefix added to order numbers (e.g., TK-12345)</p>
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Low Stock Threshold</label>
            <input
              type="number"
              defaultValue="10"
              min="0"
              className="w-full px-3 py-2.5 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
              disabled={isPending}
            />
            <p className="mt-1 text-xs text-muted-foreground">Products below this quantity will be flagged as low stock</p>
          </div>
          <div className="flex items-center justify-between py-2">
            <div>
              <p className="text-sm font-medium text-foreground">Auto-cancel unpaid orders</p>
              <p className="text-xs text-muted-foreground">Automatically cancel orders after 48 hours if unpaid</p>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input type="checkbox" defaultChecked className="sr-only peer" disabled={isPending} />
              <div className="w-11 h-6 bg-muted peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-primary-500 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary-500"></div>
            </label>
          </div>
        </div>
      </div>

      <div className="flex justify-end">
        <button
          disabled={isPending}
          className="px-6 py-2.5 rounded-lg bg-primary-500 text-white font-medium text-sm hover:bg-primary-600 disabled:opacity-50 transition-colors"
        >
          {isPending ? 'Saving...' : 'Save Changes'}
        </button>
      </div>
    </div>
  );
}

function AppearanceSettings() {
  const [isPending, startTransition] = useTransition();
  const [theme, setTheme] = useState('system');

  return (
    <div className="space-y-6">
      <div className="rounded-xl border border-border bg-card p-6">
        <h3 className="text-lg font-semibold text-foreground mb-4">Theme</h3>
        <div className="grid grid-cols-3 gap-4">
          {[
            { id: 'light', label: 'Light', icon: '☀️' },
            { id: 'dark', label: 'Dark', icon: '🌙' },
            { id: 'system', label: 'System', icon: '💻' },
          ].map((option) => (
            <button
              key={option.id}
              onClick={() => setTheme(option.id)}
              className={cn(
                'p-4 rounded-lg border-2 transition-colors text-center',
                theme === option.id
                  ? 'border-primary-500 bg-primary-500/10'
                  : 'border-border hover:border-primary-300'
              )}
              disabled={isPending}
            >
              <span className="text-2xl">{option.icon}</span>
              <p className="text-sm font-medium text-foreground mt-2">{option.label}</p>
            </button>
          ))}
        </div>
      </div>

      <div className="rounded-xl border border-border bg-card p-6">
        <h3 className="text-lg font-semibold text-foreground mb-4">Sidebar</h3>
        <div className="space-y-4">
          <div className="flex items-center justify-between py-2">
            <div>
              <p className="text-sm font-medium text-foreground">Compact sidebar</p>
              <p className="text-xs text-muted-foreground">Use a narrower sidebar with icons only</p>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input type="checkbox" className="sr-only peer" disabled={isPending} />
              <div className="w-11 h-6 bg-muted peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-primary-500 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary-500"></div>
            </label>
          </div>
          <div className="flex items-center justify-between py-2">
            <div>
              <p className="text-sm font-medium text-foreground">Fixed header</p>
              <p className="text-xs text-muted-foreground">Keep the header visible when scrolling</p>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input type="checkbox" defaultChecked className="sr-only peer" disabled={isPending} />
              <div className="w-11 h-6 bg-muted peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-primary-500 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary-500"></div>
            </label>
          </div>
        </div>
      </div>

      <div className="flex justify-end">
        <button
          disabled={isPending}
          className="px-6 py-2.5 rounded-lg bg-primary-500 text-white font-medium text-sm hover:bg-primary-600 disabled:opacity-50 transition-colors"
        >
          {isPending ? 'Saving...' : 'Save Changes'}
        </button>
      </div>
    </div>
  );
}

function NotificationSettings() {
  const [isPending, startTransition] = useTransition();

  const notifications = [
    { id: 'new_order', label: 'New orders', description: 'Get notified when a new order is placed', email: true, push: true },
    { id: 'low_stock', label: 'Low stock alerts', description: 'Get notified when products are running low', email: true, push: false },
    { id: 'new_user', label: 'New user registrations', description: 'Get notified when a new user signs up', email: false, push: true },
    { id: 'refund_request', label: 'Refund requests', description: 'Get notified when a refund is requested', email: true, push: true },
    { id: 'system_alert', label: 'System alerts', description: 'Critical system notifications', email: true, push: true },
  ];

  return (
    <div className="space-y-6">
      <div className="rounded-xl border border-border bg-card p-6">
        <h3 className="text-lg font-semibold text-foreground mb-4">Notification Preferences</h3>
        <div className="space-y-4">
          {notifications.map((notif) => (
            <div key={notif.id} className="flex items-center justify-between py-3 border-b border-border last:border-0">
              <div>
                <p className="text-sm font-medium text-foreground">{notif.label}</p>
                <p className="text-xs text-muted-foreground">{notif.description}</p>
              </div>
              <div className="flex items-center gap-4">
                <label className="flex items-center gap-2">
                  <input type="checkbox" defaultChecked={notif.email} className="w-4 h-4 rounded border-border text-primary-500 focus:ring-primary-500" disabled={isPending} />
                  <span className="text-xs text-muted-foreground">Email</span>
                </label>
                <label className="flex items-center gap-2">
                  <input type="checkbox" defaultChecked={notif.push} className="w-4 h-4 rounded border-border text-primary-500 focus:ring-primary-500" disabled={isPending} />
                  <span className="text-xs text-muted-foreground">Push</span>
                </label>
              </div>
            </div>
          ))}
        </div>
      </div>

      <div className="flex justify-end">
        <button
          disabled={isPending}
          className="px-6 py-2.5 rounded-lg bg-primary-500 text-white font-medium text-sm hover:bg-primary-600 disabled:opacity-50 transition-colors"
        >
          {isPending ? 'Saving...' : 'Save Changes'}
        </button>
      </div>
    </div>
  );
}

function SecuritySettings() {
  const [isPending, startTransition] = useTransition();

  return (
    <div className="space-y-6">
      <div className="rounded-xl border border-border bg-card p-6">
        <h3 className="text-lg font-semibold text-foreground mb-4">Two-Factor Authentication</h3>
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm text-foreground">Enable 2FA for your account</p>
            <p className="text-xs text-muted-foreground">Add an extra layer of security to your account</p>
          </div>
          <button className="px-4 py-2 rounded-lg border border-border bg-card text-foreground font-medium text-sm hover:bg-muted transition-colors">
            Setup 2FA
          </button>
        </div>
      </div>

      <div className="rounded-xl border border-border bg-card p-6">
        <h3 className="text-lg font-semibold text-foreground mb-4">Change Password</h3>
        <div className="space-y-4 max-w-md">
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Current Password</label>
            <input
              type="password"
              className="w-full px-3 py-2.5 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
              disabled={isPending}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">New Password</label>
            <input
              type="password"
              className="w-full px-3 py-2.5 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
              disabled={isPending}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Confirm New Password</label>
            <input
              type="password"
              className="w-full px-3 py-2.5 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500"
              disabled={isPending}
            />
          </div>
          <button
            disabled={isPending}
            className="px-4 py-2 rounded-lg bg-primary-500 text-white font-medium text-sm hover:bg-primary-600 disabled:opacity-50 transition-colors"
          >
            Update Password
          </button>
        </div>
      </div>

      <div className="rounded-xl border border-border bg-card p-6">
        <h3 className="text-lg font-semibold text-foreground mb-4">Active Sessions</h3>
        <div className="space-y-3">
          <div className="flex items-center justify-between p-3 rounded-lg border border-border bg-muted/20">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-success-100 dark:bg-success-900/30 flex items-center justify-center">
                <svg className="w-5 h-5 text-success-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                </svg>
              </div>
              <div>
                <p className="text-sm font-medium text-foreground">Current Session</p>
                <p className="text-xs text-muted-foreground">Chrome on Ubuntu • Hanoi, Vietnam</p>
              </div>
            </div>
            <span className="px-2 py-1 rounded text-xs font-medium bg-success-100 text-success-700 dark:bg-success-900/30 dark:text-success-400">Active</span>
          </div>
        </div>
      </div>
    </div>
  );
}

function ApiSettings() {
  const [isPending, startTransition] = useTransition();

  return (
    <div className="space-y-6">
      <div className="rounded-xl border border-border bg-card p-6">
        <h3 className="text-lg font-semibold text-foreground mb-4">API Keys</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Use API keys to authenticate requests to the Tiki API. Keep your keys secure and never share them publicly.
        </p>
        <div className="space-y-3">
          <div className="flex items-center justify-between p-4 rounded-lg border border-border bg-muted/20">
            <div>
              <p className="text-sm font-medium text-foreground">Production Key</p>
              <p className="text-xs text-muted-foreground font-mono">tk_live_****************************</p>
              <p className="text-xs text-muted-foreground mt-1">Created on Jan 15, 2025</p>
            </div>
            <div className="flex items-center gap-2">
              <button className="px-3 py-1.5 rounded-md border border-border bg-card text-foreground text-xs font-medium hover:bg-muted transition-colors">
                Regenerate
              </button>
              <button className="px-3 py-1.5 rounded-md bg-danger-500 text-white text-xs font-medium hover:bg-danger-600 transition-colors">
                Revoke
              </button>
            </div>
          </div>
        </div>
        <button className="mt-4 px-4 py-2 rounded-lg border border-dashed border-border bg-card text-foreground font-medium text-sm hover:bg-muted transition-colors">
          + Generate New API Key
        </button>
      </div>

      <div className="rounded-xl border border-border bg-card p-6">
        <h3 className="text-lg font-semibold text-foreground mb-4">Webhooks</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Configure webhook endpoints to receive real-time notifications about events.
        </p>
        <div className="space-y-3">
          <div className="flex items-center justify-between p-4 rounded-lg border border-border bg-muted/20">
            <div>
              <p className="text-sm font-medium text-foreground">Order Events</p>
              <p className="text-xs text-muted-foreground font-mono">https://api.tiki.vn/webhooks/orders</p>
            </div>
            <span className="px-2 py-1 rounded text-xs font-medium bg-success-100 text-success-700 dark:bg-success-900/30 dark:text-success-400">Active</span>
          </div>
        </div>
        <button className="mt-4 px-4 py-2 rounded-lg border border-dashed border-border bg-card text-foreground font-medium text-sm hover:bg-muted transition-colors">
          + Add Webhook Endpoint
        </button>
      </div>
    </div>
  );
}

function BillingSettings() {
  return (
    <div className="space-y-6">
      <div className="rounded-xl border border-border bg-card p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-foreground">Current Plan</h3>
          <span className="px-3 py-1 rounded-full text-xs font-medium bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400">Enterprise</span>
        </div>
        <div className="grid grid-cols-3 gap-4 mb-6">
          <div className="p-4 rounded-lg border border-border bg-muted/20">
            <p className="text-2xl font-bold text-foreground">Unlimited</p>
            <p className="text-xs text-muted-foreground">Products</p>
          </div>
          <div className="p-4 rounded-lg border border-border bg-muted/20">
            <p className="text-2xl font-bold text-foreground">Unlimited</p>
            <p className="text-xs text-muted-foreground">Orders/month</p>
          </div>
          <div className="p-4 rounded-lg border border-border bg-muted/20">
            <p className="text-2xl font-bold text-foreground">24/7</p>
            <p className="text-xs text-muted-foreground">Support</p>
          </div>
        </div>
        <button className="px-4 py-2 rounded-lg bg-primary-500 text-white font-medium text-sm hover:bg-primary-600 transition-colors">
          Upgrade Plan
        </button>
      </div>

      <div className="rounded-xl border border-border bg-card p-6">
        <h3 className="text-lg font-semibold text-foreground mb-4">Payment Method</h3>
        <div className="flex items-center justify-between p-4 rounded-lg border border-border bg-muted/20">
          <div className="flex items-center gap-3">
            <div className="w-12 h-8 rounded bg-muted flex items-center justify-center text-xs font-bold">VISA</div>
            <div>
              <p className="text-sm font-medium text-foreground">•••• •••• •••• 4242</p>
              <p className="text-xs text-muted-foreground">Expires 12/2026</p>
            </div>
          </div>
          <button className="text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400">Edit</button>
        </div>
      </div>

      <div className="rounded-xl border border-border bg-card p-6">
        <h3 className="text-lg font-semibold text-foreground mb-4">Billing History</h3>
        <div className="space-y-2">
          {[
            { date: 'Jun 1, 2025', amount: '$299.00', status: 'Paid' },
            { date: 'May 1, 2025', amount: '$299.00', status: 'Paid' },
            { date: 'Apr 1, 2025', amount: '$299.00', status: 'Paid' },
          ].map((invoice, i) => (
            <div key={i} className="flex items-center justify-between py-2 border-b border-border last:border-0">
              <div>
                <p className="text-sm text-foreground">{invoice.date}</p>
                <p className="text-xs text-muted-foreground">Enterprise Plan - Monthly</p>
              </div>
              <div className="flex items-center gap-3">
                <span className="text-sm font-medium text-foreground">{invoice.amount}</span>
                <span className="px-2 py-0.5 rounded text-xs font-medium bg-success-100 text-success-700 dark:bg-success-900/30 dark:text-success-400">{invoice.status}</span>
                <button className="text-xs text-primary-600 hover:text-primary-700 dark:text-primary-400">Download</button>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
