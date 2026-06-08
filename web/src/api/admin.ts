import apiClient from './client';
import type { Product } from '../types/catalog';
import type { Order } from '../types/order';
import type { EvalRun, MetricsResult } from '../types/admin';

// Products CRUD
export async function createProduct(product: Partial<Product>): Promise<Product> {
  const { data } = await apiClient.post('/admin/products', product);
  return data.data;
}

export async function updateProduct(id: number, product: Partial<Product>): Promise<Product> {
  const { data } = await apiClient.put(`/admin/products/${id}`, product);
  return data.data;
}

export async function deleteProduct(id: number): Promise<void> {
  await apiClient.delete(`/admin/products/${id}`);
}

export async function reindexSearch(): Promise<void> {
  await apiClient.post('/admin/search/index');
}

// Admin Orders
export async function listAllOrders(page = 1, pageSize = 20, status?: string): Promise<{ orders: Order[]; total: number; page: number; page_size: number }> {
  const params: Record<string, string | number> = { page, page_size: pageSize };
  if (status) params.status = status;
  const { data } = await apiClient.get('/admin/orders', { params });
  return data.data;
}

export async function getAdminOrder(id: number): Promise<Order> {
  const { data } = await apiClient.get(`/admin/orders/${id}`);
  return data.data;
}

// Metrics
export async function getMetrics(): Promise<MetricsResult> {
  const { data } = await apiClient.get('/metrics');
  return data.data;
}

// Eval Runs
export async function listEvalRuns(category?: string, limit = 50): Promise<EvalRun[]> {
  const params: Record<string, string | number> = { limit };
  if (category) params.category = category;
  const { data } = await apiClient.get('/eval/runs', { params });
  return data.data.eval_runs;
}

export async function saveEvalRun(run: Omit<EvalRun, 'id' | 'created_at'>): Promise<EvalRun> {
  const { data } = await apiClient.post('/eval/runs', run);
  return data.data;
}
