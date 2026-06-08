export interface CheckoutInput {
  success_url: string;
  cancel_url: string;
  idempotency_key?: string;
}

export interface CheckoutResult {
  checkout_url: string;
}
