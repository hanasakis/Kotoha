import apiClient from './client';
import type { CartData, AddItemReq, UpdateQtyReq } from '../types/cart';

export async function getCart(): Promise<CartData> {
  const { data } = await apiClient.get('/cart');
  return data.data;
}

export async function addItem(input: AddItemReq): Promise<CartData> {
  const { data } = await apiClient.post('/cart/items', input);
  return data.data;
}

export async function updateQty(skuId: number, input: UpdateQtyReq): Promise<CartData> {
  const { data } = await apiClient.put(`/cart/items/${skuId}`, input);
  return data.data;
}

export async function removeItem(skuId: number): Promise<CartData> {
  const { data } = await apiClient.delete(`/cart/items/${skuId}`);
  return data.data;
}
