import type { OrderStatus } from '../../types/order'

const statusMap: Record<OrderStatus, { label: string; color: string }> = {
  pending_payment: { label: '待支付', color: 'bg-yellow-100 text-yellow-800' },
  paid: { label: '已支付', color: 'bg-blue-100 text-blue-800' },
  shipped: { label: '已发货', color: 'bg-purple-100 text-purple-800' },
  delivered: { label: '已送达', color: 'bg-green-100 text-green-800' },
  cancelled: { label: '已取消', color: 'bg-gray-100 text-gray-600' },
  refunded: { label: '已退款', color: 'bg-orange-100 text-orange-800' },
  partially_refunded: { label: '部分退款', color: 'bg-orange-100 text-orange-800' },
  expired: { label: '已过期', color: 'bg-gray-100 text-gray-600' },
  payment_failed: { label: '支付失败', color: 'bg-red-100 text-red-800' },
}

export default function OrderStatusBadge({ status }: { status: OrderStatus }) {
  const s = statusMap[status] || { label: status, color: 'bg-gray-100' }
  return (
    <span className={`inline-block px-2 py-0.5 rounded-full text-xs font-medium ${s.color}`}>
      {s.label}
    </span>
  )
}
