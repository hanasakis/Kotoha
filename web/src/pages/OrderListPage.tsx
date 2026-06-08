import { useState, useEffect, useCallback } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate, Link } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import * as orderApi from '../api/order'
import * as paymentApi from '../api/payment'
import PriceDisplay from '../components/shared/PriceDisplay'
import OrderStatusBadge from '../components/shared/OrderStatusBadge'
import type { Order } from '../types/order'

const PAYMENT_WINDOW_MS = 10 * 60 * 1000 // 10 minutes

function useCountdown(deadline: number) {
  const [left, setLeft] = useState(Math.max(0, deadline - Date.now()))

  useEffect(() => {
    const tick = () => setLeft(Math.max(0, deadline - Date.now()))
    tick()
    const id = setInterval(tick, 1000)
    return () => clearInterval(id)
  }, [deadline])

  return left
}

function PayButton({ order }: { order: Order }) {
  const locale = useAuthStore((s) => s.locale)
  const navigate = useNavigate()
  const [paying, setPaying] = useState(false)

  const deadline = new Date(order.created_at).getTime() + PAYMENT_WINDOW_MS
  const left = useCountdown(deadline)

  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)

  const handlePay = async (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setPaying(true)
    try {
      const base = window.location.origin
      const result = await paymentApi.createCheckout(order.id, {
        success_url: `${base}/orders/${order.id}/success`,
        cancel_url: `${base}/orders/${order.id}/cancel`,
      })
      window.location.href = result.checkout_url
    } catch {
      setPaying(false)
    }
  }

  if (left <= 0) return null

  const mins = Math.floor(left / 60000)
  const secs = Math.floor((left % 60000) / 1000)

  return (
    <button
      onClick={handlePay}
      disabled={paying}
      className="px-4 py-1.5 bg-rose-500 text-white text-xs rounded-full hover:bg-rose-600 disabled:opacity-50 font-medium"
    >
      {paying
        ? t('跳转中...', 'Redirecting...')
        : `${t('继续支付', 'Pay Now')} ${mins}:${String(secs).padStart(2, '0')}`}
    </button>
  )
}

export default function OrderListPage() {
  const locale = useAuthStore((s) => s.locale)
  const [page, setPage] = useState(1)

  const { data, isLoading } = useQuery({
    queryKey: ['orders', page],
    queryFn: () => orderApi.listOrders(page, 10),
  })

  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)
  const totalPages = data ? Math.ceil(data.total / data.page_size) : 0

  return (
    <div className="max-w-3xl mx-auto px-4 py-6">
      <h1 className="text-2xl font-bold mb-6">{t('我的订单', 'My Orders')}</h1>

      {isLoading ? (
        <div className="text-gray-400 text-sm">{t('加载中...', 'Loading...')}</div>
      ) : (data?.orders.length || 0) === 0 ? (
        <div className="text-center py-16 text-gray-400">
          <p className="mb-4">{t('暂无订单', 'No orders yet')}</p>
          <Link to="/products" className="text-rose-500 hover:underline">
            {t('去逛逛', 'Browse Products')} →
          </Link>
        </div>
      ) : (
        <>
          <div className="space-y-3">
            {(data?.orders || []).map((o) => (
              <Link
                key={o.id}
                to={`/orders/${o.id}`}
                className="block p-4 bg-white border rounded-xl hover:shadow-sm transition-shadow"
              >
                <div className="flex justify-between items-start mb-2">
                  <div>
                    <p className="text-xs text-gray-400">#{o.order_no}</p>
                    <p className="text-sm text-gray-500 mt-1">
                      {new Date(o.created_at).toLocaleDateString(locale === 'zh' ? 'zh-CN' : 'en-US')}
                    </p>
                  </div>
                  <OrderStatusBadge status={o.status} />
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-xs text-gray-400">
                    {o.items?.length || 0} {t('件商品', ' items')}
                  </span>
                  <div className="flex items-center gap-3">
                    <PriceDisplay price={o.total_amount} className="font-bold" />
                    {o.status === 'pending_payment' && <PayButton order={o} />}
                  </div>
                </div>
              </Link>
            ))}
          </div>

          {totalPages > 1 && (
            <div className="flex justify-center gap-2 mt-6">
              <button
                onClick={() => setPage(Math.max(1, page - 1))}
                disabled={page <= 1}
                className="px-3 py-1.5 border rounded text-sm disabled:opacity-30"
              >
                {t('上一页', 'Prev')}
              </button>
              <span className="px-3 py-1.5 text-sm text-gray-500">
                {page}/{totalPages}
              </span>
              <button
                onClick={() => setPage(Math.min(totalPages, page + 1))}
                disabled={page >= totalPages}
                className="px-3 py-1.5 border rounded text-sm disabled:opacity-30"
              >
                {t('下一页', 'Next')}
              </button>
            </div>
          )}
        </>
      )}
    </div>
  )
}
