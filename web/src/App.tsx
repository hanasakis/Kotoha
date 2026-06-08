import { Routes, Route } from 'react-router-dom'
import { useEffect, useRef } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useAuthStore } from './store/authStore'
import { useCartStore } from './store/cartStore'
import AppLayout from './components/layout/AppLayout'
import AuthGuard from './components/shared/AuthGuard'
import AdminGuard from './components/layout/AdminGuard'
import AdminLayout from './components/layout/AdminLayout'

import HomePage from './pages/HomePage'
import SearchPage from './pages/SearchPage'
import ProductListPage from './pages/ProductListPage'
import ProductDetailPage from './pages/ProductDetailPage'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import ForgotPasswordPage from './pages/ForgotPasswordPage'
import ResetPasswordPage from './pages/ResetPasswordPage'
import CartPage from './pages/CartPage'
import CheckoutPage from './pages/CheckoutPage'
import OrderListPage from './pages/OrderListPage'
import OrderDetailPage from './pages/OrderDetailPage'
import OrderSuccessPage from './pages/OrderSuccessPage'
import OrderCancelPage from './pages/OrderCancelPage'
import ProfilePage from './pages/ProfilePage'
import AddressPage from './pages/AddressPage'
import AddressFormPage from './pages/AddressFormPage'
import AgentChatPage from './pages/AgentChatPage'
import NotFoundPage from './pages/NotFoundPage'
import AdminDashboardPage from './pages/admin/AdminDashboardPage'
import AdminProductsPage from './pages/admin/AdminProductsPage'
import AdminOrdersPage from './pages/admin/AdminOrdersPage'
import AdminEvaluationsPage from './pages/admin/AdminEvaluationsPage'
import AdminObservabilityPage from './pages/admin/AdminObservabilityPage'

export default function App() {
  const hydrate = useAuthStore((s) => s.hydrate)
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const user = useAuthStore((s) => s.user)
  const fetchCart = useCartStore((s) => s.fetchCart)
  const queryClient = useQueryClient()
  const prevUserId = useRef<number | null>(null)

  useEffect(() => {
    hydrate()
  }, [hydrate])

  // Clear user-scoped queries when switching users (prevents cross-user data leak)
  useEffect(() => {
    const currentId = user?.id ?? null
    if (prevUserId.current !== null && currentId !== prevUserId.current) {
      // Only clear user-specific queries — keep public data (catalog, categories, products)
      queryClient.removeQueries({ predicate: (query) => {
        const key = query.queryKey[0] as string
        return ['orders', 'order', 'profile', 'preference', 'addresses', 'cart', 'admin-orders-count', 'admin-eval-runs', 'admin-metrics'].includes(key)
      }})
    }
    prevUserId.current = currentId
  }, [user?.id, queryClient])

  useEffect(() => {
    if (isAuthenticated) fetchCart()
  }, [isAuthenticated, fetchCart])

  return (
    <Routes>
      <Route element={<AppLayout />}>
        <Route path="/" element={<HomePage />} />
        <Route path="/search" element={<SearchPage />} />
        <Route path="/products" element={<ProductListPage />} />
        <Route path="/products/:id" element={<ProductDetailPage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/forgot-password" element={<ForgotPasswordPage />} />
        <Route path="/reset-password" element={<ResetPasswordPage />} />

        <Route path="/orders/:id/success" element={<OrderSuccessPage />} />
        <Route path="/orders/:id/cancel" element={<OrderCancelPage />} />

        <Route element={<AuthGuard />}>
          <Route path="/cart" element={<CartPage />} />
          <Route path="/checkout" element={<CheckoutPage />} />
          <Route path="/orders" element={<OrderListPage />} />
          <Route path="/orders/:id" element={<OrderDetailPage />} />
          <Route path="/profile" element={<ProfilePage />} />
          <Route path="/profile/addresses" element={<AddressPage />} />
          <Route path="/profile/addresses/new" element={<AddressFormPage />} />
          <Route path="/profile/addresses/:id/edit" element={<AddressFormPage />} />
          <Route path="/agent" element={<AgentChatPage />} />
        </Route>

        <Route element={<AdminGuard />}>
          <Route element={<AdminLayout />}>
            <Route path="/admin" element={<AdminDashboardPage />} />
            <Route path="/admin/products" element={<AdminProductsPage />} />
            <Route path="/admin/orders" element={<AdminOrdersPage />} />
            <Route path="/admin/evaluations" element={<AdminEvaluationsPage />} />
            <Route path="/admin/observability" element={<AdminObservabilityPage />} />
          </Route>
        </Route>

        <Route path="*" element={<NotFoundPage />} />
      </Route>
    </Routes>
  )
}
