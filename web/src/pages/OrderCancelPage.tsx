import { useParams, Link } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'

export default function OrderCancelPage() {
  const { id } = useParams<{ id: string }>()
  const locale = useAuthStore((s) => s.locale)
  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)

  return (
    <div className="max-w-md mx-auto mt-16 px-4 text-center">
      <div className="text-6xl mb-4">💳</div>
      <h1 className="text-2xl font-bold mb-4">{t('支付已取消', 'Payment Cancelled')}</h1>
      <p className="text-gray-500 mb-6">
        {t('您的支付已被取消，订单仍然保留，您可以稍后继续支付。', 'Your payment was cancelled. The order is still saved — you can pay later.')}
      </p>
      <div className="flex justify-center gap-4">
        <Link
          to={`/orders/${id}`}
          className="px-6 py-2 bg-rose-500 text-white rounded-lg hover:bg-rose-600"
        >
          {t('查看订单', 'View Order')}
        </Link>
        <Link
          to="/orders"
          className="px-6 py-2 border rounded-lg hover:bg-gray-50"
        >
          {t('我的订单', 'My Orders')}
        </Link>
      </div>
    </div>
  )
}
