import apiClient from './client';
import type { CheckoutInput, CheckoutResult } from '../types/payment';

export async function createCheckout(orderId: number, input: CheckoutInput): Promise<CheckoutResult> {
  const { data } = await apiClient.post(`/orders/${orderId}/checkout`, input);
  return data.data;
}

// Public sync — no auth required, uses order_no as secret
export async function syncPaymentPublic(orderNo: string): Promise<{ status: string }> {
  const { data } = await apiClient.post('/public/sync-payment', { order_no: orderNo });
  return data.data;
}
