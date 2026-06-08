import apiClient from './client';
import type { Category, Product, PaginatedProducts, SearchResult, SKU } from '../types/catalog';

export async function getCategories(): Promise<Category[]> {
  const { data } = await apiClient.get('/catalog/categories');
  return data.data;
}

export async function getProducts(page = 1, pageSize = 20, keyword?: string, categoryId?: number): Promise<PaginatedProducts> {
  const params: Record<string, string | number | undefined> = { page, page_size: pageSize, keyword };
  if (categoryId) params.category_id = categoryId;
  const { data } = await apiClient.get('/catalog/products', { params });
  return data.data;
}

export async function getProduct(id: number): Promise<Product> {
  const { data } = await apiClient.get(`/catalog/products/${id}`);
  return data.data;
}

export async function searchProducts(query: string, limit = 20): Promise<SearchResult> {
  const { data } = await apiClient.get('/catalog/search', { params: { q: query, limit } });
  return data.data;
}
