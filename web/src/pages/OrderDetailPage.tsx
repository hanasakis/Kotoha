import { useState, useEffect } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useParams, Link } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import * as orderApi from '../api/order'
import * as paymentApi from '../api/payment'
import PriceDisplay from '../components/shared/PriceDisplay'
import OrderStatusBadge from '../components/shared/OrderStatusBadge'

const PAYMENT_WINDOW_MS = 10 * 60 * 1000

export default function OrderDetailPage() {
  const { id } = useParams<{ id: string }>()
  const locale = useAuthStore((s) => s.locale)
  const queryClient = useQueryClient()
  const [cancelling, setCancelling] = useState(false)
  const [paying, setPaying] = useState(false)
  const [syncing, setSyncing] = useState(false)

  const { data: order, isLoading } = useQuery({
    queryKey: ['order', id],
    queryFn: () => orderApi.getOrder(Number(id)),
    enabled: !!id,
    refetchInterval: false,
  })

  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)

  const deadline = order ? new Date(order.created_at).getTime() + PAYMENT_WINDOW_MS : 0
  const [left, setLeft] = useState(0)
  useEffect(() => {
    if (!order) return
    const tick = () => setLeft(Math.max(0, deadline - Date.now()))
    tick()
    const interval = setInterval(tick, 1000)
    return () => clearInterval(interval)
  }, [deadline, order])

  // Auto-sync payment for pending_payment orders (in case webhook was missed)
  useEffect(() => {
    if (!order || order.status !== 'pending_payment') return
    // Sync once on mount in the background
    const sync = async () => {
      try {
        const result = await orderApi.syncPayment(order.id)
        if (result.status !== 'pending_payment') {
          queryClient.invalidateQueries({ queryKey: ['order', id] })
        }
      } catch { /* silent — sync is best-effort */ }
    }
    sync()
  }, [order?.id]) // eslint-disable-line react-hooks/exhaustive-deps

  const handlePay = async () => {
    if (!order) return
    setPaying(true)
    try {
      const base = window.location.origin
      const result = await paymentApi.createCheckout(order.id, {
        success_url: `${base}/orders/${order.id}/success?order_no=${encodeURIComponent(order.order_no)}`,
        cancel_url: `${base}/orders/${order.id}/cancel`,
      })
      window.location.href = result.checkout_url
    } catch {
      setPaying(false)
    }
  }

  const handleSyncPayment = async () => {
    if (!order) return
    setSyncing(true)
    try {
      await orderApi.syncPayment(order.id)
      queryClient.invalidateQueries({ queryKey: ['order', id] })
    } catch { /* ignore */ }
    setSyncing(false)
  }

  const handleCancel = async () => {
    if (!order) return
    if (!confirm(t('确定取消此订单？', 'Cancel this order?'))) return
    setCancelling(true)
    try {
      await orderApi.cancelOrder(order.id)
      queryClient.invalidateQueries({ queryKey: ['order', id] })
    } finally {
      setCancelling(false)
    }
  }

  if (isLoading) {
    return <div className="max-w-3xl mx-auto px-4 py-8 text-gray-400">{t('加载中...', 'Loading...')}</div>
  }

  if (!order) {
    return (
      <div className="max-w-3xl mx-auto px-4 py-16 text-center text-gray-400">
        {t('订单不存在', 'Order not found')}
      </div>
    )
  }

  const canCancel = order.status === 'pending_payment'
  const canPay = order.status === 'pending_payment' && left > 0
  const canSync = order.status === 'pending_payment'

  return (
    <div className="max-w-3xl mx-auto px-4 py-6">
      <Link to="/orders" className="text-sm text-gray-400 hover:text-gray-600 mb-4 inline-block">
        ← {t('返回订单列表', 'Back to Orders')}
      </Link>

      <div className="flex justify-between items-start mb-6">
        <div>
          <h1 className="text-2xl font-bold">{t('订单详情', 'Order Detail')}</h1>
          <p className="text-sm text-gray-400 mt-1">#{order.order_no}</p>
        </div>
        <OrderStatusBadge status={order.status} />
      </div>

      {/* Payment countdown */}
      {order.status === 'pending_payment' && left > 0 && (
        <div className="mb-4 p-3 bg-yellow-50 border border-yellow-200 rounded-lg text-sm text-yellow-800">
          {t('剩余支付时间', 'Payment window')}: {Math.floor(left / 60000)}:
          {String(Math.floor((left % 60000) / 1000)).padStart(2, '0')}
        </div>
      )}

      {/* Already paid banner */}
      {order.status === 'paid' && (
        <div className="mb-4 p-3 bg-green-50 border border-green-200 rounded-lg text-sm text-green-800">
          {t('订单已支付，等待发货', 'Order paid, awaiting shipment')}
        </div>
      )}

      {/* Items */}
      <div className="border rounded-xl divide-y mb-6">
        {(order.items || []).map((item) => (
          <div key={item.id} className="flex justify-between p-3 text-sm">
            <span>{item.name} x{item.quantity}</span>
            <PriceDisplay price={item.price * item.quantity} />
          </div>
        ))}
      </div>

      {/* Summary */}
      <div className="flex justify-between items-center text-lg font-bold mb-6">
        <span>{t('合计', 'Total')}</span>
        <PriceDisplay price={order.total_amount} className="text-rose-500" />
      </div>

      {/* Info */}
      <div className="text-sm text-gray-500 space-y-1 mb-6">
        <p>{t('下单时间', 'Created')}: {new Date(order.created_at).toLocaleString(locale === 'zh' ? 'zh-CN' : 'en-US')}</p>
        {order.paid_at && (
          <p>{t('支付时间', 'Paid')}: {new Date(order.paid_at).toLocaleString(locale === 'zh' ? 'zh-CN' : 'en-US')}</p>
        )}
      </div>

      <div className="flex gap-3 flex-wrap">
        {canPay && (
          <button
            onClick={handlePay}
            disabled={paying}
            className="px-6 py-2 bg-rose-500 text-white rounded-lg text-sm hover:bg-rose-600 disabled:opacity-50 font-medium"
          >
            {paying ? t('跳转中...', 'Redirecting...') : t('继续支付', 'Pay Now')}
          </button>
        )}
        {canSync && (
          <button
            onClick={handleSyncPayment}
            disabled={syncing}
            className="px-4 py-2 border border-blue-300 text-blue-600 rounded-lg text-sm hover:bg-blue-50 disabled:opacity-50"
          >
            {syncing ? t('同步中...', 'Syncing...') : t('刷新支付状态', 'Refresh Payment Status')}
          </button>
        )}
        {canCancel && (
          <button
            onClick={handleCancel}
            disabled={cancelling}
            className="px-4 py-2 border border-red-300 text-red-500 rounded-lg text-sm hover:bg-red-50 disabled:opacity-50"
          >
            {cancelling ? t('取消中...', 'Cancelling...') : t('取消订单', 'Cancel Order')}
          </button>
        )}
      </div>
    </div>
  )
}
