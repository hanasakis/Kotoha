import { create } from 'zustand';
import type { CartItem } from '../types/cart';
import * as cartApi from '../api/cart';

interface CartState {
  items: CartItem[];
  loading: boolean;

  fetchCart: () => Promise<void>;
  addItem: (skuId: number, quantity: number) => Promise<void>;
  updateQty: (skuId: number, quantity: number) => Promise<void>;
  removeItem: (skuId: number) => Promise<void>;
  clearCart: () => void;
}

export const useCartStore = create<CartState>((set) => ({
  items: [],
  loading: false,

  fetchCart: async () => {
    set({ loading: true });
    try {
      const data = await cartApi.getCart();
      set({ items: data.items });
    } catch {
      set({ items: [] });
    } finally {
      set({ loading: false });
    }
  },

  addItem: async (skuId: number, quantity: number) => {
    const data = await cartApi.addItem({ sku_id: skuId, quantity });
    set({ items: data.items });
  },

  updateQty: async (skuId: number, quantity: number) => {
    const data = await cartApi.updateQty(skuId, { quantity });
    set({ items: data.items });
  },

  removeItem: async (skuId: number) => {
    const data = await cartApi.removeItem(skuId);
    set({ items: data.items });
  },

  clearCart: () => set({ items: [] }),
}));
