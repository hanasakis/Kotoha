# Kotoha Backend Requirements

> **Date:** 2026-06-04 | **Status:** Living document

## 1. Project Overview

Kotoha (果多哈) is an AI-powered snack e-commerce platform. The backend is a modular monolith built in Go providing user authentication, product catalog browsing, semantic/hybrid search, shopping cart, Stripe checkout, order lifecycle management, refunds, and an LLM agent for conversational shopping assistance.

### 1.1 Tech Stack

| Layer | Technology |
|-------|------------|
| Language | Go 1.24 |
| HTTP Framework | Gin |
| ORM | GORM |
| Primary DB | PostgreSQL |
| Cache / Session | Redis |
| Vector Search | Milvus (dense embedding + BM25 hybrid) |
| Embeddings | Ollama (bge-large-zh-v1.5) |
| LLM | Ollama (qwen3.5:9b) |
| Payments | Stripe (Checkout Sessions + Refunds) |
| Email | Resend SMTP |
| Observability | Langfuse (traces + scores) |

### 1.2 Package Map

```
internal/
  agent/          LLM agent with tool-use execution
  auth/           Registration, login, JWT, password reset, deletion
  cart/           Redis-backed shopping cart
  catalog/        Product/category CRUD, seed data
  config/         Structured config via viper + .env
  i18n/           JSON locale bundles (zh, en)
  middleware/      CORS, auth, rate limiting, request ID
  observability/  Metrics (Redis counters), eval run persistence, Langfuse
  order/          Order lifecycle, status management, refunds
  payment/        Stripe webhook handling, checkout creation
  router/         Route registration, DI wiring
  search/         Dense + BM25 hybrid search over Milvus
  user/           Profile, addresses, preferences
pkg/
  db/             Auto-migration
  embedding/      Ollama embedding HTTP client
  langfuse/       Langfuse trace + score API client
  log/            Structured logging
  mail/           SMTP client (Resend)
  milvus/         Milvus REST API v2 client (collection, insert, search, query)
  ollama/         Ollama chat + embedding HTTP client
  redis/          Redis client wrapper
  stripe/         Stripe checkout + refund API client
```

---

## 2. Functional Requirements

### FR-1: Authentication & Session Management

**FR-1.1 Registration**
- Accept email + password (bcrypt hash), create user with `customer` role
- Reject duplicate emails with i18n error `auth.email_exists`
- Return JWT access + refresh tokens on success

**FR-1.2 Login**
- Validate email + bcrypt password
- Issue access token (short-lived, default 15min) + refresh token (long-lived, default 30d)
- Sessions persisted in DB, refresh rotation supported

**FR-1.3 Token Refresh**
- `POST /api/v1/auth/refresh` accepts a refresh token, rotates it (old marked used, new issued)

**FR-1.4 Logout**
- Protected endpoint; marks the session as used, invalidating the refresh token

**FR-1.5 Password Reset**
- `POST /api/v1/auth/forgot-password`: always returns 200 (anti-enumeration); if email exists, generates 32-byte crypto/rand hex token, stores SHA256(token) in DB with 1-hour expiry, sends email via Resend SMTP
- `POST /api/v1/auth/reset-password`: validates token (SHA256 match, not expired, not used), bcrypt-hashes new password, marks token used, invalidates all user sessions
- Frontend reset URL: `https://kotoha.shop/reset-password?token=<token>`

**FR-1.6 Account Deletion (Soft Delete)**
- `DELETE /api/v1/auth/account`: requires password confirmation, invalidates all sessions, GORM soft-delete on user row
- Orders, addresses, preferences preserved (only user row affected)
- Login after deletion returns `auth.login_failed`

### FR-2: User Profile Management

**FR-2.1 Profile**
- `GET /api/v1/profile`, `PUT /api/v1/profile` — read/write nickname, avatar, phone

**FR-2.2 Addresses**
- Full CRUD: `GET/POST /api/v1/addresses`, `PUT/DELETE /api/v1/addresses/:id`
- Fields: receiver name, phone, province, city, district, detail, is_default

**FR-2.3 Preferences**
- `GET/PUT /api/v1/preferences` — locale, currency, dietary preferences (JSON)

### FR-3: Product Catalog

**FR-3.1 Categories**
- `GET /api/v1/catalog/categories` — list all product categories

**FR-3.2 Products**
- `GET /api/v1/catalog/products` — paginated list with category filter
- `GET /api/v1/catalog/products/:id` — product detail including SKUs
- Admin-only CRUD: `POST/PUT/DELETE /api/v1/admin/products`

**FR-3.3 Seed Data**
- On startup, catalog service seeds default categories + products if DB is empty
- Product images served via S3-compatible storage public URL

### FR-4: Hybrid Search

**FR-4.1 Dense Vector Search**
- Product fields (name, name_en, description text) embedded via Ollama
- Milvus collection with HNSW index, configurable ef/m parameters

**FR-4.2 BM25 Keyword Search**
- Milvus dynamic field storage for text keywords
- Client-side keyword matching via Milvus entity query with `like` filter

**FR-4.3 Hybrid Fusion**
- Client-side weighted score fusion: `combined = DenseWeight * norm_score + BM25Weight * position_score`
- Dense scores min-max normalized; BM25 scores are rank-based positional (1/rank)
- Configurable weights (default: dense 0.65, BM25 0.35), recall top-K
- Silent fallback to dense-only when BM25 query fails

**FR-4.4 Indexing**
- `POST /api/v1/admin/search/index` — re-index all products into Milvus

### FR-5: Shopping Cart

**FR-5.1 Cart Operations (Redis-backed, per-user)**
- `GET /api/v1/cart` — get cart contents with product details
- `POST /api/v1/cart/items` — add SKU with quantity; validates stock
- `PUT /api/v1/cart/items/:skuID` — update quantity
- `DELETE /api/v1/cart/items/:skuID` — remove item

**FR-5.2 Inventory Validation**
- On add/update, verify SKU exists and stock >= requested quantity
- Return `cart.insufficient_stock` or `cart.sku_not_found` on failure

### FR-6: Orders & Checkout

**FR-6.1 Order Creation**
- `POST /api/v1/orders` — creates order from cart items, clears cart on success
- Computes totals from SKU prices; validates stock at creation time

**FR-6.2 Order Listing**
- `GET /api/v1/orders` — user's order history
- `GET /api/v1/orders/:id` — order detail with items

**FR-6.3 Checkout**
- `POST /api/v1/orders/:id/checkout` — creates Stripe Checkout Session
- Returns Stripe checkout URL for redirect
- Order status: `pending` → Stripe session created

**FR-6.4 Order Cancellation**
- `POST /api/v1/orders/:id/cancel` — cancels order
- Status `pending`: restore stock, mark `cancelled`
- Status `paid`: create Stripe refund, restore stock, mark `refunded` or `partially_refunded`
- Other statuses: return `order.cannot_cancel`

**FR-6.5 Order Status Machine**
```
pending → [checkout session created]
pending → cancelled (user/expiry)
pending → paid (checkout.session.completed)
paid → refunded / partially_refunded (charge.refunded)
pending → expired (checkout.session.expired)
pending → payment_failed (payment_intent.payment_failed)
```

### FR-7: Payment & Webhooks

**FR-7.1 Stripe Webhook**
- `POST /api/v1/webhook` — receives Stripe events, validates signature

**FR-7.2 Supported Webhook Events**
- `checkout.session.completed` — mark order paid, create payment record
- `checkout.session.async_payment_succeeded` — same handling as completed
- `checkout.session.expired` — mark order expired
- `checkout.session.async_payment_failed` — mark order payment_failed
- `payment_intent.payment_failed` — mark order payment_failed
- `charge.refunded` — update payment refund fields, set order to refunded/partially_refunded

### FR-8: AI Shopping Agent

**FR-8.1 Conversational Interface**
- `POST /api/v1/agent/chat` — accepts user message + optional conversation history
- Returns agent reply (tool result or chat response)

**FR-8.2 Tool Execution**
- LLM outputs structured JSON tool calls (no free-text)
- Supported tools: `search_products`, `get_product`, `list_categories`, `add_to_cart`, `get_cart`, `remove_from_cart`
- Executor runs tool against real services (catalog, search, cart)
- Results fed back to LLM for final response generation

**FR-8.3 Langfuse Tracing**
- Each chat interaction traced in Langfuse with input, output, tool calls, and metadata

### FR-9: Observability & Metrics

**FR-9.1 Real-Time Metrics (Redis Counters)**
- Search count, average relevance score
- Cart add count
- Order created count
- `GET /api/v1/metrics` returns structured metric categories

**FR-9.2 Evaluation Dashboard**
- `POST /api/v1/eval/runs` — persist eval run (category, total/passed/failed, accuracy, metrics JSON)
- `GET /api/v1/eval/runs?category=&limit=` — list historical eval runs
- Eval results pushed to Langfuse as scores for centralized observability

---

## 3. Non-Functional Requirements

### NFR-1: Security

- **NFR-1.1** Passwords hashed with bcrypt, never logged or returned in API responses
- **NFR-1.2** JWT access tokens short-lived (15min default), refresh tokens rotated on use
- **NFR-1.3** Password reset tokens use SHA256 storage (not plaintext), 1-hour expiry
- **NFR-1.4** Anti-email-enumeration: forgot-password endpoint always returns 200
- **NFR-1.5** Stripe webhook signature verification before processing
- **NFR-1.6** Rate limiting on protected routes (120 req/min, burst 20)
- **NFR-1.7** CORS configuration per environment
- **NFR-1.8** All secrets from environment variables or `.env` file; config validates JWT_SECRET at startup

### NFR-2: Reliability

- **NFR-2.1** Milvus collection init failure is non-fatal (retries on first search)
- **NFR-2.2** Catalog seed data failure is non-fatal (logs warning)
- **NFR-2.3** BM25 keyword search failure silently falls back to dense-only
- **NFR-2.4** Stripe SDK with built-in retry logic
- **NFR-2.5** Langfuse push is fire-and-forget (async flush, no blocking)

### NFR-3: Performance

- **NFR-3.1** Cart operations via Redis (sub-millisecond reads/writes)
- **NFR-3.2** DB connection pooling (configurable max open/idle connections)
- **NFR-3.3** Hybrid search recall top-K capped, final sort client-side for small result sets

### NFR-4: Observability

- **NFR-4.1** Request ID injected on all requests for log correlation
- **NFR-4.2** Structured logging via pkg/log
- **NFR-4.3** Langfuse traces for agent chat interactions
- **NFR-4.4** Langfuse scores for evaluation results
- **NFR-4.5** Redis-backed atomic counters for business metrics

### NFR-5: Internationalization

- **NFR-5.1** All API error codes returned as i18n keys (e.g., `"auth.invalid_token"`)
- **NFR-5.2** Supported locales: `zh` (default), `en`
- **NFR-5.3** Locale resolved from `Accept-Language` header via middleware

### NFR-6: Code Quality

- **NFR-6.1** Modular monolith: each `internal/<domain>` has clean boundaries (handler → service → repository)
- **NFR-6.2** Cross-cutting concerns via function callbacks, not direct package imports (avoids circular deps)
- **NFR-6.3** No package-level init() for side effects; explicit initialization in router.Setup
- **NFR-6.4** Config struct-based, no global config variables

---

## 4. API Route Summary

### Public (no auth)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| POST | `/api/v1/auth/register` | Register |
| POST | `/api/v1/auth/login` | Login |
| POST | `/api/v1/auth/refresh` | Refresh token |
| POST | `/api/v1/auth/forgot-password` | Request password reset |
| POST | `/api/v1/auth/reset-password` | Reset password with token |
| GET | `/api/v1/catalog/categories` | List categories |
| GET | `/api/v1/catalog/products` | List products |
| GET | `/api/v1/catalog/products/:id` | Product detail |
| GET | `/api/v1/catalog/search` | Search products (keyword + semantic) |
| POST | `/api/v1/webhook` | Stripe webhook |

### Protected (requires JWT)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/logout` | Logout |
| DELETE | `/api/v1/auth/account` | Delete account |
| GET/PUT | `/api/v1/profile` | Read/update profile |
| GET | `/api/v1/addresses` | List addresses |
| POST | `/api/v1/addresses` | Create address |
| PUT | `/api/v1/addresses/:id` | Update address |
| DELETE | `/api/v1/addresses/:id` | Delete address |
| GET/PUT | `/api/v1/preferences` | Read/update preferences |
| GET | `/api/v1/cart` | Get cart |
| POST | `/api/v1/cart/items` | Add to cart |
| PUT | `/api/v1/cart/items/:skuID` | Update cart item qty |
| DELETE | `/api/v1/cart/items/:skuID` | Remove from cart |
| POST | `/api/v1/orders` | Create order |
| GET | `/api/v1/orders` | List orders |
| GET | `/api/v1/orders/:id` | Order detail |
| POST | `/api/v1/orders/:id/cancel` | Cancel order |
| POST | `/api/v1/orders/:id/checkout` | Create Stripe checkout |
| POST | `/api/v1/agent/chat` | AI agent chat |
| GET | `/api/v1/metrics` | Business metrics |
| GET | `/api/v1/eval/runs` | List eval runs |
| POST | `/api/v1/eval/runs` | Save eval run |

### Admin (requires JWT + admin role)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/admin/products` | Create product |
| PUT | `/api/v1/admin/products/:id` | Update product |
| DELETE | `/api/v1/admin/products/:id` | Delete product |
| POST | `/api/v1/admin/search/index` | Re-index all products |

---

## 5. External Integrations

| Service | Purpose | Configuration |
|---------|---------|---------------|
| PostgreSQL | Primary data store | `DB_*` env vars |
| Redis | Cart storage, session, metrics counters | `REDIS_*` env vars |
| Milvus | Vector + keyword search | `MILVUS_*` env vars |
| Ollama | LLM chat + text embeddings | `OLLAMA_HOST` |
| Stripe | Payment checkout + refunds + webhooks | `STRIPE_*` env vars |
| Resend (SMTP) | Password reset emails | `SMTP_*` env vars |
| Langfuse | LLM observability (traces + scores) | `LANGFUSE_*` env vars |
| S3-compatible | Product image hosting | `S3_*` env vars |

---

## 6. Future Work (Backlog)

- OpenAPI / Swagger specification generation
- MRR/NDCG computation for search evaluation
- CI-integrated evaluation pipeline
- Langfuse trace expansion for non-agent flows
- Stripe AI agent, Tax, and Radar integration
- Architecture diagrams (C4 model)
- Product review/rating system
- Order shipment tracking
