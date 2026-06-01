// Core Types - Enterprise Ecommerce Platform
export interface User {
  id: string;
  email: string;
  username: string;
  display_name: string;
  phone: string;
  avatar_url: string;
  status: "active" | "inactive" | "locked";
  role: string;
  created_at: string;
  updated_at: string;
}

export interface Category {
  id: string;
  parent_id?: string | null;
  name: string;
  slug: string;
  description?: string;
  image_url?: string;
  sort_order: number;
  is_active: boolean;
  depth?: number;
  path?: string;
  children?: Category[];
  product_count?: number;
  created_at?: string;
}

export interface Product {
  id: string;
  tiki_product_id?: string;
  shop_id: string;
  category_id: string;
  category_name?: string;
  name: string;
  description?: string;
  short_description?: string;
  sku_code?: string;
  brand?: string;
  image_url: string;
  thumbnail_url?: string;
  images?: ProductImage[];
  price: number;
  original_price?: number | null;
  discount_percent?: number | null;
  stock: number;
  sold_count: number;
  quantity_sold_text?: string;
  rating_average?: number | null;
  rating_count?: number;
  review_count?: number;
  seller_name?: string;
  seller_avatar_url?: string;
  is_tiki_trading?: boolean;
  is_official?: boolean;
  is_sponsored?: boolean;
  status: "draft" | "active" | "inactive" | "out_of_stock";
  condition: string;
  weight?: number;
  dimensions?: string;
  created_at: string;
  updated_at: string;
}

export interface ProductImage {
  id: string;
  url: string;
  thumbnail_url?: string;
  alt_text?: string;
  sort_order: number;
  is_primary: boolean;
}

export interface ProductVariant {
  id: string;
  name: string;
  price: number;
  stock: number;
  attributes: Record<string, string>;
}

export interface CartItem {
  id: string;
  product_id: string;
  sku_id: string;
  name: string;
  image_url: string;
  price: number;
  original_price?: number;
  quantity: number;
  stock: number;
  is_selected: boolean;
  sku_name?: string;
  shop_id?: string;
  shop_name?: string;
}

export interface Cart {
  id: string;
  user_id?: string;
  session_id?: string;
  items: CartItem[];
  total_items: number;
  subtotal: number;
  discount: number;
  shipping_fee: number;
  total: number;
  currency: string;
}

export interface Order {
  id: string;
  order_number: string;
  user_id: string;
  status: OrderStatus;
  payment_status: string;
  shipping_address: Address;
  billing_address?: Address;
  subtotal: number;
  shipping_fee: number;
  discount: number;
  total: number;
  currency: string;
  payment_method?: string;
  note?: string;
  items: OrderItem[];
  timeline?: OrderTimelineItem[];
  created_at: string;
  updated_at: string;
}

export type OrderStatus =
  | "pending"
  | "confirmed"
  | "processing"
  | "shipped"
  | "delivered"
  | "cancelled"
  | "refunded";

export interface OrderItem {
  id: string;
  product_id: string;
  sku_id: string;
  name: string;
  image_url: string;
  price: number;
  quantity: number;
  total: number;
}

export interface OrderTimelineItem {
  status: string;
  timestamp: string;
  description: string;
}

export interface Address {
  id?: string;
  name: string;
  phone: string;
  address_line1: string;
  address_line2?: string;
  city: string;
  state: string;
  postal_code: string;
  country: string;
  is_default?: boolean;
}

export interface Customer {
  id: string;
  email: string;
  username: string;
  display_name: string;
  phone: string;
  avatar_url: string;
  status: string;
  total_orders: number;
  total_spent: number;
  fraud_score?: number;
  last_login_at?: string;
  created_at: string;
}

// API Response Types
export interface ApiResponse<T> {
  success: boolean;
  data: T;
  message?: string;
  error_code?: string;
  trace_id?: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

// Dashboard Types
export interface DashboardMetrics {
  gmv: number;
  gmv_change: number;
  revenue: number;
  revenue_change: number;
  orders: number;
  orders_change: number;
  customers: number;
  customers_change: number;
  conversion_rate: number;
  aov: number;
  active_users: number;
}

export interface ChartDataPoint {
  label: string;
  value: number;
  timestamp?: string;
  category?: string;
}

export interface Alert {
  id: string;
  type: "info" | "warning" | "error" | "success";
  severity: "low" | "medium" | "high" | "critical";
  title: string;
  message: string;
  source?: string;
  is_read: boolean;
  created_at: string;
}

export interface NotificationItem {
  id: string;
  type: string;
  title: string;
  message: string;
  link?: string;
  is_read: boolean;
  created_at: string;
}

export interface ProductListResponse {
  products: Product[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface OrderListResponse {
  orders: Order[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}
export interface SearchFilters {
  query?: string;
  category_id?: string;
  min_price?: number;
  max_price?: number;
  brand?: string[];
  rating?: number;
  sort?: string;
  page?: number;
  limit?: number;
}

export interface SearchResult {
  products: Product[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
  filters?: {
    brands: { name: string; count: number }[];
    price_ranges: { min: number; max: number; count: number }[];
    categories: { id: string; name: string; count: number }[];
  };
}

// Store Types
export interface CartStore {
  items: CartItem[];
  isLoading: boolean;
  addItem: (item: Omit<CartItem, "id" | "is_selected"> & { is_selected?: boolean }) => Promise<void>;
  removeItem: (id: string) => Promise<void>;
  updateQuantity: (id: string, quantity: number) => Promise<void>;
  toggleSelect: (id: string) => void;
  selectAll: () => void;
  clearCart: () => void;
  getSelectedItems: () => CartItem[];
  getSubtotal: () => number;
  getTotal: () => number;
}

export interface AuthStore {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<{ access_token: string; refresh_token: string; expires_in: number; token_type: string; session_id: string }>;
  logout: () => Promise<void>;
  register: (data: RegisterRequest) => Promise<{ access_token: string; refresh_token: string; expires_in: number; token_type: string; session_id: string }>;
  refreshUser: () => Promise<void>;
}

export interface RegisterRequest {
  email: string;
  username: string;
  password: string;
  confirm_password: string;
  display_name?: string;
}

export interface UIStore {
  sidebarOpen: boolean;
  mobileMenuOpen: boolean;
  theme: "light" | "dark";
  toastNotifications: NotificationItem[];
  toggleSidebar: () => void;
  toggleMobileMenu: () => void;
  setTheme: (theme: "light" | "dark") => void;
  addToast: (toast: Omit<NotificationItem, "id" | "created_at" | "is_read">) => void;
  dismissToast: (id: string) => void;
}
