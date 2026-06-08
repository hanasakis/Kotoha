export interface CartItem {
  sku_id: number;
  product_id: number;
  product_name: string;
  sku_name: string;
  price: number;
  quantity: number;
  image_url: string;
  stock: number;
}

export interface CartData {
  items: CartItem[];
}

export interface AddItemReq {
  sku_id: number;
  quantity: number;
}

export interface UpdateQtyReq {
  quantity: number;
}
