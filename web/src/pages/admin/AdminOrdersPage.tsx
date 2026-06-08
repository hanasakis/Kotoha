import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useAuthStore } from '../../store/authStore'
import * as adminApi from '../../api/admin'
import PriceDisplay from '../../components/shared/PriceDisplay'
import OrderStatusBadge from '../../components/shared/OrderStatusBadge'
import type { Order } from '../../types/order'

const orderStatuses = [
  { value: '', label: 'All', labelZh: '全部' },
  { value: 'pending_payment', label: 'Pending', labelZh: '待支付' },
  { value: 'paid', label: 'Paid', labelZh: '已支付' },
  { value: 'shipped', label: 'Shipped', labelZh: '已发货' },
  { value: 'delivered', label: 'Delivered', labelZh: '已送达' },
  { value: 'cancelled', label: 'Cancelled', labelZh: '已取消' },
  { value: 'refunded', label: 'Refunded', labelZh: '已退款' },
  { value: 'expired', label: 'Expired', labelZh: '已过期' },
]

function OrderDetailModal({ order, onClose }: { order: Order; onClose: () => void }) {
  const locale = useAuthStore((s) => s.locale)
  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)

  return (
    <div className="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-lg max-h-[90vh] overflow-y-auto">
        <div className="px-6 py-4 border-b flex justify-between items-center">
          <h3 className="font-semibold">{t('订单详情', 'Order Detail')} #{order.order_no}</h3>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 text-lg">&times;</button>
        </div>
        <div className="px-6 py-4 space-y-4">
          <div className="grid grid-cols-2 gap-3 text-sm">
            <div>
              <span className="text-gray-400">{t('订单号', 'Order No')}</span>
              <p className="font-mono text-xs mt-0.5">{order.order_no}</p>
            </div>
            <div>
              <span className="text-gray-400">{t('状态', 'Status')}</span>
              <p className="mt-0.5"><OrderStatusBadge status={order.status} /></p>
            </div>
            <div>
              <span className="text-gray-400">{t('用户ID', 'User ID')}</span>
              <p className="mt-0.5">{order.user_id}</p>
            </div>
            <div>
              <span className="text-gray-400">{t('总金额', 'Total')}</span>
              <p className="mt-0.5 font-semibold"><PriceDisplay price={order.total_amount} /></p>
            </div>
          </div>
          {order.stripe_session_id && (
            <div className="text-sm">
              <span className="text-gray-400">{t('Stripe Session', 'Stripe Session')}</span>
              <p className="font-mono text-xs mt-0.5 break-all">{order.stripe_session_id}</p>
            </div>
          )}
          <div className="text-sm">
            <span className="text-gray-400">{t('创建时间', 'Created')}</span>
            <p className="mt-0.5">{new Date(order.created_at).toLocaleString()}</p>
          </div>
          {order.paid_at && (
            <div className="text-sm">
              <span className="text-gray-400">{t('支付时间', 'Paid')}</span>
              <p className="mt-0.5">{new Date(order.paid_at).toLocaleString()}</p>
            </div>
          )}

          {/* Order Items */}
          <div>
            <h4 className="text-sm font-medium text-gray-700 mb-2">{t('订单商品', 'Items')}</h4>
            <table className="w-full text-sm border rounded-lg overflow-hidden">
              <thead>
                <tr className="bg-gray-50 text-left text-gray-500">
                  <th className="py-2 px-3 font-medium">{t('商品', 'Item')}</th>
                  <th className="py-2 px-3 font-medium text-center">{t('单价', 'Price')}</th>
                  <th className="py-2 px-3 font-medium text-center">{t('数量', 'Qty')}</th>
                  <th className="py-2 px-3 font-medium text-right">{t('小计', 'Subtotal')}</th>
                </tr>
              </thead>
              <tbody>
                {(order.items || []).map((item) => (
                  <tr key={item.id} className="border-t">
                    <td className="py-2 px-3">{item.name}</td>
                    <td className="py-2 px-3 text-center"><PriceDisplay price={item.price} /></td>
                    <td className="py-2 px-3 text-center">{item.quantity}</td>
                    <td className="py-2 px-3 text-right"><PriceDisplay price={item.price * item.quantity} /></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  )
}

export default function AdminOrdersPage() {
  const locale = useAuthStore((s) => s.locale)
  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)
  const [page, setPage] = useState(1)
  const [status, setStatus] = useState('')
  const [selectedOrder, setSelectedOrder] = useState<Order | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['admin-orders', page, status],
    queryFn: () => adminApi.listAllOrders(page, 20, status || undefined),
  })

  return (
    <div>
      <h1 className="text-xl font-bold text-gray-900 mb-6">{t('订单管理', 'Orders')}</h1>

      {/* Status filter */}
      <div className="flex gap-2 mb-4 flex-wrap">
        {orderStatuses.map((s) => (
          <button
            key={s.value}
            onClick={() => { setStatus(s.value); setPage(1) }}
            className={`px-3 py-1 rounded-full text-xs transition-colors ${
              status === s.value
                ? 'bg-rose-500 text-white'
                : 'bg-white border text-gray-500 hover:border-rose-300'
            }`}
          >
            {t(s.labelZh, s.label)}
          </button>
        ))}
      </div>

      <div className="bg-white border rounded-xl overflow-hidden">
        {isLoading ? (
          <div className="p-8 text-center text-gray-400">{t('加载中...', 'Loading...')}</div>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-gray-500 border-b bg-gray-50">
                <th className="py-3 px-4 font-medium">{t('订单号', 'Order No')}</th>
                <th className="py-3 px-4 font-medium">{t('用户', 'User')}</th>
                <th className="py-3 px-4 font-medium">{t('金额', 'Amount')}</th>
                <th className="py-3 px-4 font-medium">{t('状态', 'Status')}</th>
                <th className="py-3 px-4 font-medium">{t('时间', 'Date')}</th>
                <th className="py-3 px-4 font-medium">{t('操作', 'Actions')}</th>
              </tr>
            </thead>
            <tbody>
              {(data?.orders || []).map((o) => (
                <tr key={o.id} className="border-t hover:bg-gray-50">
                  <td className="py-2.5 px-4 font-mono text-xs">{o.order_no}</td>
                  <td className="py-2.5 px-4 text-gray-500">#{o.user_id}</td>
                  <td className="py-2.5 px-4 font-medium"><PriceDisplay price={o.total_amount} /></td>
                  <td className="py-2.5 px-4"><OrderStatusBadge status={o.status} /></td>
                  <td className="py-2.5 px-4 text-gray-400 text-xs">{new Date(o.created_at).toLocaleDateString()}</td>
                  <td className="py-2.5 px-4">
                    <button onClick={() => setSelectedOrder(o)} className="text-rose-500 hover:underline text-xs">
                      {t('详情', 'Detail')}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}

        {data && data.total > 20 && (
          <div className="px-4 py-3 border-t flex justify-between items-center text-sm text-gray-500">
            <span>{t(`共 ${data.total} 单`, `${data.total} orders total`)}</span>
            <div className="flex gap-2">
              <button onClick={() => setPage((p) => Math.max(1, p - 1))} disabled={page <= 1} className="px-3 py-1 border rounded hover:bg-gray-100 disabled:opacity-30">
                {t('上一页', 'Prev')}
              </button>
              <span className="px-3 py-1">{page}</span>
              <button onClick={() => setPage((p) => p + 1)} disabled={page * 20 >= data.total} className="px-3 py-1 border rounded hover:bg-gray-100 disabled:opacity-30">
                {t('下一页', 'Next')}
              </button>
            </div>
          </div>
        )}
      </div>

      {selectedOrder && (
        <OrderDetailModal order={selectedOrder} onClose={() => setSelectedOrder(null)} />
      )}
    </div>
  )
}
