export type OrderStatus =
  | 'pending_payment'
  | 'paid'
  | 'shipped'
  | 'delivered'
  | 'cancelled'
  | 'refunded'
  | 'partially_refunded'
  | 'expired'
  | 'payment_failed';

export interface OrderItem {
  id: number;
  order_id: number;
  product_id: number;
  sku_id: number;
  name: string;
  price: number;
  quantity: number;
}

export interface Order {
  id: number;
  user_id: number;
  order_no: string;
  status: OrderStatus;
  total_amount: number;
  currency: string;
  stripe_session_id: string;
  paid_at: string | null;
  created_at: string;
  updated_at: string;
  items: OrderItem[];
}

export interface PaginatedOrders {
  orders: Order[];
  total: number;
  page: number;
  page_size: number;
}

export interface CreateOrderInput {
  address_id: number;
  note?: string;
}
