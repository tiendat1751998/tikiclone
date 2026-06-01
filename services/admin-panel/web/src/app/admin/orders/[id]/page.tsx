import { Suspense } from 'react';
import { notFound } from 'next/navigation';
import { ErrorBoundary } from '@/components/admin';
import { formatVND, formatRelativeTime } from '@shopee/shared-utils';

interface OrderDetail {
  id: string;
  order_number: string;
  status: string;
  customer: {
    id: string;
    name: string;
    email: string;
    phone: string;
  };
  shipping_address: {
    street: string;
    city: string;
    district: string;
    ward: string;
    phone: string;
  };
  items: {
    id: string;
    product_name: string;
    product_image: string;
    quantity: number;
    price: number;
    total: number;
  }[];
  subtotal: number;
  shipping_fee: number;
  discount: number;
  total: number;
  payment_method: string;
  payment_status: string;
  timeline: {
    status: string;
    timestamp: string;
    note?: string;
  }[];
  created_at: string;
  updated_at: string;
}

async function getOrderDetail(id: string) {
  const res = await fetch(
    `${process.env.API_GATEWAY_URL || 'http://localhost:8080'}/api/admin/orders/${id}`,
    {
      next: { revalidate: 30 },
      headers: { 'Content-Type': 'application/json' },
    }
  );

  if (res.status === 404) {
    notFound();
  }

  if (!res.ok) {
    throw new Error('Failed to fetch order details');
  }

  return res.json();
}

function OrderDetailLoading() {
  return <div className="h-96 bg-muted rounded-lg animate-pulse" />;
}

export default async function OrderDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <a href="/admin/orders" className="hover:text-foreground">
          Orders
        </a>
        <span>/</span>
        <span className="text-foreground">Order Details</span>
      </div>

      <ErrorBoundary>
        <Suspense fallback={<OrderDetailLoading />}>
          <OrderDetailContent id={id} />
        </Suspense>
      </ErrorBoundary>
    </div>
  );
}

async function OrderDetailContent({ id }: { id: string }) {
  let order: OrderDetail;

  try {
    const data = await getOrderDetail(id);
    order = data.order || data;
  } catch {
    return (
      <div className="text-center py-12 text-muted-foreground">
        <p>Unable to load order details. Please try again later.</p>
      </div>
    );
  }

  const statusStyles: Record<string, string> = {
    pending: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',
    confirmed: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
    processing: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400',
    shipped: 'bg-cyan-100 text-cyan-700 dark:bg-cyan-900/30 dark:text-cyan-400',
    delivered: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
    cancelled: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
    refunded: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
  };

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div className="lg:col-span-2 space-y-6">
        <div className="rounded-xl border border-border bg-card p-6">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h1 className="text-xl font-bold text-foreground">
                Order #{order.order_number}
              </h1>
              <p className="text-sm text-muted-foreground mt-1">
                Placed {formatRelativeTime(order.created_at)}
              </p>
            </div>
            <span
              className={`px-3 py-1 rounded-full text-sm font-medium capitalize ${statusStyles[order.status]}`}
            >
              {order.status}
            </span>
          </div>

          <div className="space-y-4">
            <h3 className="text-sm font-semibold text-foreground">Order Items</h3>
            <div className="divide-y divide-border">
              {order.items.map((item) => (
                <div key={item.id} className="py-4 flex items-center gap-4">
                  <div className="w-16 h-16 rounded-lg bg-muted flex items-center justify-center">
                    {item.product_image ? (
                      <img
                        src={item.product_image}
                        alt={item.product_name}
                        className="w-full h-full object-cover rounded-lg"
                      />
                    ) : (
                      <svg
                        className="w-6 h-6 text-muted-foreground"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"
                        />
                      </svg>
                    )}
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="font-medium text-foreground truncate">
                      {item.product_name}
                    </p>
                    <p className="text-sm text-muted-foreground">
                      Qty: {item.quantity} x {formatVND(item.price)}
                    </p>
                  </div>
                  <p className="font-medium text-foreground">
                    {formatVND(item.total)}
                  </p>
                </div>
              ))}
            </div>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <h3 className="text-sm font-semibold text-foreground mb-4">Order Timeline</h3>
          <div className="space-y-4">
            {order.timeline.map((event, index) => (
              <div key={index} className="flex gap-4">
                <div className="flex flex-col items-center">
                  <div className="w-3 h-3 rounded-full bg-primary-500" />
                  {index < order.timeline.length - 1 && (
                    <div className="w-0.5 h-full bg-border mt-1" />
                  )}
                </div>
                <div className="pb-4">
                  <p className="text-sm font-medium text-foreground capitalize">
                    {event.status}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {formatRelativeTime(event.timestamp)}
                  </p>
                  {event.note && (
                    <p className="text-xs text-muted-foreground mt-1">{event.note}</p>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      <div className="space-y-6">
        <div className="rounded-xl border border-border bg-card p-6">
          <h3 className="text-sm font-semibold text-foreground mb-4">Customer</h3>
          <div className="space-y-2">
            <p className="font-medium text-foreground">{order.customer.name}</p>
            <p className="text-sm text-muted-foreground">{order.customer.email}</p>
            <p className="text-sm text-muted-foreground">{order.customer.phone}</p>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <h3 className="text-sm font-semibold text-foreground mb-4">
            Shipping Address
          </h3>
          <div className="text-sm text-muted-foreground space-y-1">
            <p>{order.shipping_address.street}</p>
            <p>
              {order.shipping_address.ward}, {order.shipping_address.district}
            </p>
            <p>{order.shipping_address.city}</p>
            <p className="text-foreground mt-2">{order.shipping_address.phone}</p>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <h3 className="text-sm font-semibold text-foreground mb-4">
            Payment Summary
          </h3>
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-muted-foreground">Subtotal</span>
              <span className="text-foreground">{formatVND(order.subtotal)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">Shipping</span>
              <span className="text-foreground">{formatVND(order.shipping_fee)}</span>
            </div>
            {order.discount > 0 && (
              <div className="flex justify-between">
                <span className="text-muted-foreground">Discount</span>
                <span className="text-success-600">-{formatVND(order.discount)}</span>
              </div>
            )}
            <div className="border-t border-border pt-2 mt-2">
              <div className="flex justify-between font-semibold">
                <span className="text-foreground">Total</span>
                <span className="text-foreground">{formatVND(order.total)}</span>
              </div>
            </div>
          </div>
          <div className="mt-4 pt-4 border-t border-border">
            <p className="text-xs text-muted-foreground">
              Payment: <span className="capitalize">{order.payment_method}</span>
            </p>
            <p className="text-xs text-muted-foreground">
              Status:{' '}
              <span className="capitalize">{order.payment_status}</span>
            </p>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <h3 className="text-sm font-semibold text-foreground mb-4">
            Update Status
          </h3>
          <select className="w-full px-3 py-2 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 mb-3">
            <option value="">Select new status</option>
            <option value="confirmed">Confirmed</option>
            <option value="processing">Processing</option>
            <option value="shipped">Shipped</option>
            <option value="delivered">Delivered</option>
            <option value="cancelled">Cancelled</option>
            <option value="refunded">Refunded</option>
          </select>
          <textarea
            placeholder="Add a note (optional)"
            rows={2}
            className="w-full px-3 py-2 rounded-lg border border-border bg-card text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 resize-none mb-3"
          />
          <button className="w-full py-2 px-4 rounded-lg bg-primary-500 text-white font-medium text-sm hover:bg-primary-600 transition-colors">
            Update Order
          </button>
        </div>
      </div>
    </div>
  );
}
