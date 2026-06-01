"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { CartStore, CartItem } from "@/types";
import { cartApi } from "@/lib/api/client";
import { useAuthStore } from "./auth";

export const useCartStore = create<CartStore>()(
  persist(
    (set, get) => ({
      items: [],
      isLoading: false,

      addItem: async (item) => {
        const prev = get().items;
        set({ isLoading: true });

        const existing = prev.find(
          (i) => i.product_id === item.product_id && i.sku_id === item.sku_id
        );
        if (existing) {
          set({
            items: prev.map((i) =>
              i.id === existing.id
                ? { ...i, quantity: i.quantity + item.quantity }
                : i
            ),
          });
        } else {
          set({
            items: [
              ...prev,
              {
                id: crypto.randomUUID(),
                product_id: item.product_id,
                sku_id: item.sku_id,
                name: item.name,
                image_url: item.image_url,
                price: item.price,
                original_price: item.original_price,
                quantity: item.quantity,
                stock: item.stock || 0,
                is_selected: true,
                sku_name: item.sku_name,
                shop_id: item.shop_id,
                shop_name: item.shop_name,
              },
            ],
          });
        }

        const isAuthenticated = useAuthStore.getState().isAuthenticated;
        if (!isAuthenticated) {
          set({ isLoading: false });
          return;
        }

        try {
          const cart = await cartApi.addItem({
            product_id: item.product_id,
            sku_id: item.sku_id,
            quantity: item.quantity,
            name: item.name,
            price: item.price,
            image_url: item.image_url,
            shop_id: item.shop_id || "",
            shop_name: item.shop_name || "",
          });
          if (cart && cart.items) {
            set({ items: cart.items, isLoading: false });
          } else {
            set({ isLoading: false });
          }
        } catch {
          set({ items: prev, isLoading: false });
        }
      },

      removeItem: async (id) => {
        const prev = get().items;
        set({ items: prev.filter((i) => i.id !== id) });

        const isAuthenticated = useAuthStore.getState().isAuthenticated;
        if (!isAuthenticated) return;

        try {
          await cartApi.removeItem(id);
        } catch {
          set({ items: prev });
        }
      },

      updateQuantity: async (id, quantity) => {
        const prev = get().items;
        set({
          items: prev.map((i) =>
            i.id === id ? { ...i, quantity } : i
          ),
        });

        const isAuthenticated = useAuthStore.getState().isAuthenticated;
        if (!isAuthenticated) return;

        try {
          await cartApi.updateItem(id, quantity);
        } catch {
          set({ items: prev });
        }
      },

      toggleSelect: (id) => {
        set({
          items: get().items.map((i) =>
            i.id === id ? { ...i, is_selected: !i.is_selected } : i
          ),
        });
      },

      selectAll: () => {
        set({ items: get().items.map((i) => ({ ...i, is_selected: true })) });
      },

      clearCart: () => {
        const prev = get().items;
        set({ items: [] });

        const isAuthenticated = useAuthStore.getState().isAuthenticated;
        if (!isAuthenticated) return;

        cartApi.clear().catch(() => {
          set({ items: prev });
        });
      },

      getSelectedItems: () => get().items.filter((i) => i.is_selected),

      getSubtotal: () =>
        (get().items ?? [])
          .filter((i) => i.is_selected)
          .reduce((sum, i) => sum + i.price * i.quantity, 0),

      getTotal: () => get().getSubtotal(),
    }),
    {
      name: "tiki-cart",
      partialize: (state) => ({ items: state.items }),
    }
  )
);
