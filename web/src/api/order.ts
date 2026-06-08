import apiClient from './client';
import type { Order, PaginatedOrders, CreateOrderInput } from '../types/order';

export async function createOrder(input: CreateOrderInput): Promise<Order> {
  const { data } = await apiClient.post('/orders', input);
  return data.data;
}

export async function listOrders(page = 1, pageSize = 20): Promise<PaginatedOrders> {
  const { data } = await apiClient.get('/orders', { params: { page, page_size: pageSize } });
  return data.data;
}

export async function getOrder(id: number): Promise<Order> {
  const { data } = await apiClient.get(`/orders/${id}`);
  return data.data;
}

export async function cancelOrder(id: number): Promise<void> {
  await apiClient.post(`/orders/${id}/cancel`);
}

export async function syncPayment(id: number): Promise<{ status: string }> {
  const { data } = await apiClient.post(`/orders/${id}/sync-payment`);
  return data.data;
}
