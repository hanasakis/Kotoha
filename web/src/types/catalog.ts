export interface Category {
  id: number;
  name: string;
  name_en: string;
  parent_id: number | null;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

export interface SKU {
  id: number;
  product_id: number;
  name: string;
  price: number;
  stock: number;
  image_url: string;
  is_default: boolean;
  created_at: string;
  updated_at: string;
}

export interface Product {
  id: number;
  category_id: number;
  name: string;
  name_en: string;
  description: string;
  description_en: string;
  image_url: string;
  tags: string;
  scenes: string;
  taste: string;
  allergens: string;
  ingredients: string;
  weight_gram: number;
  shelf_days: number;
  is_active: boolean;
  click_count: number;
  buy_count: number;
  created_at: string;
  updated_at: string;
  skus: SKU[];
}

export interface PaginatedProducts {
  products: Product[];
  total: number;
  page: number;
  page_size: number;
}

export interface SearchResult {
  query: string;
  results: Product[];
}
