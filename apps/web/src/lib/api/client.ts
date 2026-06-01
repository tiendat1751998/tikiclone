import type {
  ApiResponse,
  PaginatedResponse,
  User,
  RegisterRequest,
  Product,
  ProductListResponse,
  SearchResult,
  Cart,
  Order,
  Customer,
  DashboardMetrics,
  ChartDataPoint,
  Alert,
  Address,
} from "@/types";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "/api/v1";

function getAuthToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("access_token");
}

function setAuthToken(token: string | null) {
  if (typeof window === "undefined") return;
  if (token) {
    localStorage.setItem("access_token", token);
  } else {
    localStorage.removeItem("access_token");
    localStorage.removeItem("refresh_token");
  }
}

let refreshPromise: Promise<boolean> | null = null;

class ApiError extends Error {
  status: number;
  code: string;
  traceId?: string;

  constructor(status: number, code: string, message: string, traceId?: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.traceId = traceId;
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {},
  _isRefreshRetry = false
): Promise<T> {
  const token = getAuthToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
    credentials: "include",
  });

  // Handle 401 - try refresh
  if (res.status === 401 && token) {
    // Prevent infinite retry loop
    if (_isRefreshRetry) {
      setAuthToken(null);
      throw new ApiError(401, "UNAUTHORIZED", "Session expired");
    }

    const refreshToken = localStorage.getItem("refresh_token");
    if (refreshToken) {
      try {
        // Mutex: serialize concurrent 401s to a single refresh
        if (!refreshPromise) {
          refreshPromise = fetch(`${API_BASE}/auth/refresh`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ refresh_token: refreshToken }),
          }).then(async (refreshRes) => {
            if (refreshRes.ok) {
              const refreshData = await refreshRes.json();
              const tokenData = refreshData.data || refreshData;
              setAuthToken(tokenData.access_token);
              localStorage.setItem("refresh_token", tokenData.refresh_token);
              return true;
            }
            setAuthToken(null);
            return false;
          }).catch(() => {
            setAuthToken(null);
            return false;
          });
        }

        const refreshed = await refreshPromise;
        refreshPromise = null;

        if (refreshed) {
          return request(path, options, true);
        }
      } catch {
        refreshPromise = null;
        setAuthToken(null);
      }
    }
    setAuthToken(null);
    throw new ApiError(401, "UNAUTHORIZED", "Authentication required");
  }

  let data: unknown;
  try {
    data = await res.json();
  } catch {
    throw new ApiError(res.status, "PARSE_ERROR", `Server returned ${res.status}`);
  }

  if (!res.ok) {
    throw new ApiError(
      res.status,
      (data as Record<string, string>).error_code || "UNKNOWN",
      (data as Record<string, string>).message || `Request failed with status ${res.status}`,
      (data as Record<string, string>).trace_id
    );
  }

  // Some endpoints return { data: ... } envelope, others return raw object/array
  if (data && typeof data === "object" && "data" in (data as Record<string, unknown>)) {
    return (data as Record<string, unknown>).data as T;
  }

  return data as T;
}

export const api = {
  get: <T>(path: string) => request<T>(path, { method: "GET" }),
  post: <T>(path: string, body: unknown) =>
    request<T>(path, { method: "POST", body: JSON.stringify(body) }),
  put: <T>(path: string, body: unknown) =>
    request<T>(path, { method: "PUT", body: JSON.stringify(body) }),
  patch: <T>(path: string, body: unknown) =>
    request<T>(path, { method: "PATCH", body: JSON.stringify(body) }),
  delete: <T>(path: string) => request<T>(path, { method: "DELETE" }),
};

// Auth API
export const authApi = {
  login: async (email: string, password: string) => {
    const data = await api.post<{
      access_token: string;
      refresh_token: string;
      expires_in: number;
      token_type: string;
      session_id: string;
    }>("/auth/login", { email, password });
    if (data) {
      setAuthToken(data.access_token);
      localStorage.setItem("refresh_token", data.refresh_token);
    }
    return data;
  },

  register: async (formData: {
    email: string;
    username: string;
    password: string;
    confirm_password: string;
    display_name?: string;
  }) => {
    const data = await api.post<{
      access_token: string;
      refresh_token: string;
      expires_in: number;
      token_type: string;
      session_id: string;
    }>("/auth/register", formData);
    if (data) {
      setAuthToken(data.access_token);
      localStorage.setItem("refresh_token", data.refresh_token);
    }
    return data;
  },

  logout: async () => {
    try {
      await api.post("/auth/logout", {});
    } finally {
      setAuthToken(null);
    }
  },

  me: () => api.get<import("@/types").User>("/auth/profile"),

  put: <T>(path: string, body: unknown) =>
    request<T>(path, { method: "PUT", body: JSON.stringify(body) }),

  refresh: async (refreshToken: string) => {
    const data = await api.post<{
      access_token: string;
      refresh_token: string;
      expires_in: number;
    }>("/auth/refresh", { refresh_token: refreshToken });
    if (data) {
      setAuthToken(data.access_token);
      localStorage.setItem("refresh_token", data.refresh_token);
    }
    return data;
  },
};

// Products API
export const productsApi = {
  list: (params: Record<string, string | number> = {}) =>
    api.get<import("@/types").ProductListResponse>(
      `/products?${new URLSearchParams(params as Record<string, string>)}`
    ),
  getById: (id: string) => api.get<import("@/types").Product>(`/products/${id}`),
  getFeatured: (limit = 10) =>
    api.get<import("@/types").Product[]>(`/products/featured?limit=${limit}`),
  getDeals: (limit = 20) =>
    api.get<import("@/types").Product[]>(`/products/deals?limit=${limit}`),
  getFlashSale: () =>
    api.get<{ end_time: string; products: { id: string; name: string; image_url: string; price: number; original_price: number }[] }>("/products/flash-sale"),
  getByCategory: (categoryId: string, params: Record<string, string> = {}) =>
    api.get<import("@/types").ProductListResponse>(
      `/products?category_id=${categoryId}&${new URLSearchParams(params)}`
    ),
  search: (query: string, params: Record<string, unknown> = {}) => {
    const sp = new URLSearchParams();
    sp.set("q", query);
    Object.entries(params).forEach(([k, v]) => {
      if (v !== undefined && v !== null) sp.set(k, String(v));
    });
    return api.get<import("@/types").SearchResult>(`/products/search?${sp}`);
  },
};

// Categories API
export const categoriesApi = {
  list: () => api.get<import("@/types").Category[]>("/categories"),
  getTree: () => api.get<import("@/types").Category[]>("/categories/tree"),
  getBySlug: (slug: string) =>
    api.get<import("@/types").Category>(`/categories/${slug}`),
};

// Cart API
export const cartApi = {
  get: () => api.get<import("@/types").Cart>("/cart"),
  addItem: (data: {
    product_id: string;
    sku_id: string;
    quantity: number;
    name?: string;
    price?: number;
    image_url?: string;
    shop_id?: string;
    shop_name?: string;
  }) =>
    api.post<import("@/types").Cart>("/cart/items", data),
  updateItem: (itemId: string, quantity: number) =>
    api.patch<import("@/types").Cart>(`/cart/items/${itemId}`, { quantity }),
  removeItem: (itemId: string) =>
    api.delete<import("@/types").Cart>(`/cart/items/${itemId}`),
  clear: () => api.delete("/cart"),
};

// Orders API
export const ordersApi = {
  list: (params: Record<string, string> = {}) =>
    api.get<import("@/types").PaginatedResponse<import("@/types").Order>>(
      `/orders?${new URLSearchParams(params)}`
    ),
  getById: (id: string) => api.get<import("@/types").Order>(`/orders/${id}`),
  create: (data: {
    items: { product_id: string; sku_id: string; quantity: number; price?: number; name?: string; image_url?: string; shop_id?: string; shop_name?: string }[];
    shipping_address: import("@/types").Address;
    seller_id: string;
    idempotency_key: string;
    currency: string;
    billing_address: import("@/types").Address;
    payment_method: string;
    voucher_code?: string;
  }) => api.post<import("@/types").Order>("/orders", data),
  cancel: (id: string, reason?: string) => api.post(`/orders/${id}/cancel`, { reason: reason || "User requested cancellation" }),
};

export const paymentApi = {
  authorize: (data: {
    order_id: string;
    amount: number;
    currency: string;
    payment_method: string;
    idempotency_key: string;
  }) => api.post<import("@/types").Order>("/payments", data),
  get: (id: string) => api.get(`/payments/${id}`),
  capture: (id: string) => api.post(`/payments/${id}/capture`, {}),
};

// Customers API (Admin)
export const customersApi = {
  list: (params: Record<string, string> = {}) =>
    api.get<import("@/types").PaginatedResponse<import("@/types").Customer>>(
      `/customers?${new URLSearchParams(params)}`
    ),
  getById: (id: string) => api.get<import("@/types").Customer>(`/customers/${id}`),
  update: (id: string, data: Partial<import("@/types").Customer>) =>
    api.patch<import("@/types").Customer>(`/customers/${id}`, data),
};

// Dashboard API (Admin)
export const dashboardApi = {
  getMetrics: (period: string = "7d") =>
    api.get<import("@/types").DashboardMetrics>(`/dashboard/metrics?period=${period}`),
  getRevenueChart: (period: string = "7d") =>
    api.get<import("@/types").ChartDataPoint[]>(`/dashboard/revenue?period=${period}`),
  getOrdersChart: (period: string = "7d") =>
    api.get<import("@/types").ChartDataPoint[]>(`/dashboard/orders?period=${period}`),
  getAlerts: () => api.get<import("@/types").Alert[]>("/dashboard/alerts"),
  getRealtimeStats: () =>
    api.get<{
      active_users: number;
      orders_today: number;
      revenue_today: number;
      conversion_rate: number;
    }>("/dashboard/realtime"),
};

// Promotions API
export const promotionsApi = {
  list: () => api.get("/promotions"),
  getById: (id: string) => api.get(`/promotions/${id}`),
  validate: (code: string) =>
    api.post<{ valid: boolean; discount: number; type: "percent" | "fixed" }>("/promotions/validate", { code }),
};

// Shipment API
export const shipmentsApi = {
  track: (orderId: string) =>
    api.get<{
      order_id: string;
      status: string;
      carrier: string;
      tracking_number: string;
      estimated_delivery: string;
      events: { status: string; description: string; timestamp: string; location: string }[];
    }>(`/shipments/${orderId}/track`),
};

// Recommendations API
export const recommendationsApi = {
  getRelated: (productId: string, limit = 8) =>
    api.get<Product[]>(`/recommendations?product_id=${productId}&limit=${limit}`),
};

// ============================================================
// Delivery & Geo API (delivery-tracker microservice)
// ============================================================
const DELIVERY_API_BASE =
  process.env.NEXT_PUBLIC_DELIVERY_API_URL || "";

async function deliveryRequest<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const token = getAuthToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${DELIVERY_API_BASE}${path}`, {
    ...options,
    headers,
    credentials: "include",
  });

  let data: unknown;
  try {
    data = await res.json();
  } catch {
    throw new ApiError(res.status, "PARSE_ERROR", `Server returned ${res.status}`);
  }

  if (!res.ok) {
    throw new ApiError(
      res.status,
      (data as Record<string, string>).error || "UNKNOWN",
      (data as Record<string, string>).error || `Request failed with status ${res.status}`
    );
  }

  return data as T;
}

export const deliveryApi = {
  searchAddress: (query: string) =>
    deliveryRequest<Array<{
      address: string;
      name?: string;
      lat: number;
      lng: number;
    }>>(`/api/v1/delivery/search?q=${encodeURIComponent(query)}`),

  reverseGeocode: (lat: number, lng: number) =>
    deliveryRequest<{
      address: string;
      name?: string;
      street?: string;
      city?: string;
      district?: string;
      ward?: string;
      country?: string;
      lat: number;
      lng: number;
    }>(`/api/v1/delivery/reverse?lat=${lat}&lng=${lng}`),

  calculateRoute: (pickupLat: number, pickupLng: number, dropoffLat: number, dropoffLng: number) =>
    deliveryRequest<{
      distance_meters: number;
      duration_seconds: number;
      polyline: string;
    }>("/api/v1/delivery/route", {
      method: "POST",
      body: JSON.stringify({ pickup_lat: pickupLat, pickup_lng: pickupLng, dropoff_lat: dropoffLat, dropoff_lng: dropoffLng }),
    }),

  createOrder: (data: {
    customer_id: string;
    pickup: { lat: number; lng: number; address?: string };
    dropoff: { lat: number; lng: number; address?: string };
  }) => deliveryRequest<{
    id: string;
    customer_id: string;
    status: string;
    distance_meters: number;
    duration_seconds: number;
    polyline: string;
    pickup: { lat: number; lng: number; address?: string };
    dropoff: { lat: number; lng: number; address?: string };
    created_at: string;
  }>("/api/v1/orders", { method: "POST", body: JSON.stringify(data) }),

  getOrder: (orderId: string) =>
    deliveryRequest<{
      id: string;
      customer_id: string;
      driver_id?: string;
      status: string;
      distance_meters: number;
      duration_seconds: number;
      polyline: string;
      pickup: { lat: number; lng: number; address?: string };
      dropoff: { lat: number; lng: number; address?: string };
      created_at: string;
      updated_at: string;
      assigned_at?: string;
      delivered_at?: string;
    }>(`/api/v1/orders/${orderId}`),

  getTracking: (orderId: string) =>
    deliveryRequest<Array<{
      id: string;
      order_id: string;
      driver_id: string;
      status: string;
      lat: number;
      lng: number;
      created_at: string;
    }>>(`/api/v1/orders/${orderId}/tracking`),
};

export { ApiError, setAuthToken };
