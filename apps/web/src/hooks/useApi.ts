"use client";

import {
  useQuery,
  useMutation,
  useQueryClient,
  useInfiniteQuery,
} from "@tanstack/react-query";
import { productsApi, categoriesApi, cartApi, ordersApi, authApi, customersApi, dashboardApi } from "@/lib/api/client";
import { useCartStore, useAuthStore } from "@/stores";
import type { Product, Category, CartItem, Order, Customer, SearchFilters } from "@/types";

export const queryKeys = {
  products: "products",
  product: (id: string) => ["products", id],
  categories: "categories",
  categoryTree: ["categories", "tree"],
  category: (slug: string) => ["categories", slug],
  deals: ["products", "deals"],
  featured: ["products", "featured"],
  search: (filters: SearchFilters) => ["products", "search", filters],
  cart: "cart",
  orders: "orders",
  order: (id: string) => ["orders", id],
  customers: "customers",
  customer: (id: string) => ["customers", id],
  dashboard: ["dashboard"],
  metrics: (period: string) => ["dashboard", "metrics", period],
};

export function useProducts(filters: SearchFilters = {}) {
  return useQuery({
    queryKey: [queryKeys.products, filters],
    queryFn: () => productsApi.list(filters as unknown as Record<string, string>),
    staleTime: 60000,
  });
}

export function useProduct(id: string) {
  return useQuery({
    queryKey: queryKeys.product(id),
    queryFn: () => productsApi.getById(id),
    enabled: !!id,
    staleTime: 300000,
  });
}

export function useFeaturedProducts(limit = 10) {
  return useQuery({
    queryKey: [queryKeys.featured, limit],
    queryFn: () => productsApi.getFeatured(limit),
    staleTime: 300000,
  });
}

export function useDeals(limit = 20) {
  return useQuery({
    queryKey: [queryKeys.deals, limit],
    queryFn: () => productsApi.getDeals(limit),
    staleTime: 300000,
  });
}

export function useProductSearch(filters: SearchFilters) {
  return useQuery({
    queryKey: queryKeys.search(filters),
    queryFn: () => productsApi.search(filters.query || "", filters as unknown as Record<string, unknown>),
    enabled: !!filters.query || !!filters.category_id,
    staleTime: 60000,
  });
}

export function useCategories() {
  return useQuery({
    queryKey: [queryKeys.categories],
    queryFn: () => categoriesApi.list(),
    staleTime: 3600000,
  });
}

export function useCategoryTree() {
  return useQuery({
    queryKey: queryKeys.categoryTree,
    queryFn: () => categoriesApi.getTree(),
    staleTime: 3600000,
  });
}

export function useCategory(slug: string) {
  return useQuery({
    queryKey: queryKeys.category(slug),
    queryFn: () => categoriesApi.getBySlug(slug),
    enabled: !!slug,
    staleTime: 3600000,
  });
}

export function useCart() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  return useQuery({
    queryKey: [queryKeys.cart],
    queryFn: () => cartApi.get(),
    enabled: isAuthenticated,
    staleTime: 30000,
  });
}

export function useAddToCart() {
  const queryClient = useQueryClient();
  const cartAddItem = useCartStore((s) => s.addItem);
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  return useMutation({
    mutationFn: (item: Omit<CartItem, "id">) => cartAddItem(item),
    onSuccess: () => {
      // Only invalidate cart query if user is authenticated (guest cart is local-only)
      if (isAuthenticated) {
        queryClient.invalidateQueries({ queryKey: [queryKeys.cart] });
      }
    },
  });
}

export function useOrders(params: Record<string, string> = {}) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  return useQuery({
    queryKey: [queryKeys.orders, params],
    queryFn: () => ordersApi.list(params),
    enabled: isAuthenticated,
    staleTime: 60000,
  });
}

export function useOrder(id: string) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  return useQuery({
    queryKey: queryKeys.order(id),
    queryFn: () => ordersApi.getById(id),
    enabled: !!id && isAuthenticated,
    staleTime: 30000,
  });
}

export function useCreateOrder() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ordersApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [queryKeys.orders] });
      queryClient.invalidateQueries({ queryKey: [queryKeys.cart] });
    },
  });
}

export function useUser() {
  const user = useAuthStore((s) => s.user);
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  return useQuery({
    queryKey: ["auth", "user"],
    queryFn: () => authApi.me(),
    enabled: isAuthenticated,
    initialData: user ?? undefined,
    staleTime: Infinity,
    retry: false,
  });
}

export function useLogin() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ email, password }: { email: string; password: string }) =>
      useAuthStore.getState().login(email, password),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["auth"] });
    },
  });
}

export function useLogout() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => useAuthStore.getState().logout(),
    onSuccess: () => {
      queryClient.clear();
    },
  });
}

export function useAdminCustomers(params: Record<string, string> = {}) {
  return useQuery({
    queryKey: [queryKeys.customers, params],
    queryFn: () => customersApi.list(params),
    staleTime: 60000,
  });
}

export function useAdminCustomer(id: string) {
  return useQuery({
    queryKey: queryKeys.customer(id),
    queryFn: () => customersApi.getById(id),
    enabled: !!id,
    staleTime: 60000,
  });
}

export function useDashboardMetrics(period = "7d") {
  return useQuery({
    queryKey: queryKeys.metrics(period),
    queryFn: () => dashboardApi.getMetrics(period),
    staleTime: 300000,
  });
}

export function useDashboardAlerts() {
  return useQuery({
    queryKey: [queryKeys.dashboard, "alerts"],
    queryFn: () => dashboardApi.getAlerts(),
    staleTime: 30000,
    refetchInterval: 30000,
  });
}

export function useRealtimeStats() {
  return useQuery({
    queryKey: [queryKeys.dashboard, "realtime"],
    queryFn: () => dashboardApi.getRealtimeStats(),
    staleTime: 10000,
    refetchInterval: 10000,
  });
}

// ============================================================
// Infinite scroll / load-more for product listing pages
// ============================================================

const PRODUCT_PAGE_SIZE = 20;

export interface ProductListPage {
  products: Product[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

async function fetchProductPage(
  path: string,
  page: number,
  extraParams: Record<string, string> = {}
): Promise<ProductListPage> {
  const sp = new URLSearchParams({ page: String(page), size: String(PRODUCT_PAGE_SIZE), page_size: String(PRODUCT_PAGE_SIZE), ...extraParams });
  const res = await fetch(`/api/v1${path}?${sp}`, { cache: "no-store" });
  const data = await res.json();

  if (Array.isArray(data)) {
    const products = data as Product[];
    return { products, total: products.length, page: 1, page_size: PRODUCT_PAGE_SIZE, total_pages: 1 };
  }

  const obj = data as Record<string, unknown>;
  // { data: [...], total, page, size }
  if (obj.data && Array.isArray(obj.data)) {
    return {
      products: obj.data as Product[],
      total: (obj.total as number) || 0,
      page: (obj.page as number) || page,
      page_size: (obj.size as number) || PRODUCT_PAGE_SIZE,
      total_pages: Math.ceil(((obj.total as number) || 0) / ((obj.size as number) || PRODUCT_PAGE_SIZE)) || 1,
    };
  }
  // { products: [...], total, page, size }
  if (obj.products && Array.isArray(obj.products)) {
    return {
      products: obj.products as Product[],
      total: (obj.total as number) || 0,
      page: (obj.page as number) || page,
      page_size: (obj.size as number) || PRODUCT_PAGE_SIZE,
      total_pages: Math.ceil(((obj.total as number) || 0) / ((obj.size as number) || PRODUCT_PAGE_SIZE)) || 1,
    };
  }
  // { products: [...], total, page, page_size, total_pages } (search & direct responses)
  if (obj.products || obj.data) {
    const items = (obj.products || obj.data) as Product[];
    const total = (obj.total as number) || items.length;
    const ps = (obj.page_size as number) || (obj.size as number) || PRODUCT_PAGE_SIZE;
    return {
      products: items,
      total,
      page: (obj.page as number) || page,
      page_size: ps,
      total_pages: (obj.total_pages as number) || Math.ceil(total / ps) || 1,
    };
  }
  return { products: [], total: 0, page, page_size: PRODUCT_PAGE_SIZE, total_pages: 0 };
}

export function useInfiniteProducts(
  path: string,
  extraParams: Record<string, string> = {}
) {
  return useInfiniteQuery({
    queryKey: ["products", "infinite", path, extraParams],
    queryFn: ({ pageParam }) => fetchProductPage(path, pageParam, extraParams),
    initialPageParam: 1,
    getNextPageParam: (lastPage) => {
      if (lastPage.page < lastPage.total_pages) return lastPage.page + 1;
      return undefined;
    },
    staleTime: 60000,
  });
}
