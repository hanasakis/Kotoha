import { useState, useEffect } from 'react'
import { useParams, Link, useSearchParams } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import * as orderApi from '../api/order'
import * as paymentApi from '../api/payment'

export default function OrderSuccessPage() {
  const { id } = useParams<{ id: string }>()
  const [searchParams] = useSearchParams()
  const orderNo = searchParams.get('order_no')
  const locale = useAuthStore((s) => s.locale)
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)

  const [status, setStatus] = useState<string>('syncing')

  useEffect(() => {
    // Use public endpoint with order_no (no auth required)
    if (!orderNo) {
      setStatus('pending')
      return
    }

    let attempts = 0
    const maxAttempts = 10
    const poll = async () => {
      try {
        const result = await paymentApi.syncPaymentPublic(orderNo)
        if (result.status === 'paid') {
          setStatus('paid')
          return
        }
        if (result.status === 'expired' || result.status === 'cancelled') {
          setStatus(result.status)
          return
        }
        attempts++
        if (attempts < maxAttempts) {
          setTimeout(poll, 3000)
        } else {
          setStatus('pending')
        }
      } catch {
        setStatus('pending')
      }
    }
    poll()
  }, [orderNo])

  return (
    <div className="max-w-md mx-auto mt-16 px-4 text-center">
      {status === 'paid' ? (
        <>
          <div className="text-6xl mb-4">🎉</div>
          <h1 className="text-2xl font-bold mb-4">{t('支付成功！', 'Payment Successful!')}</h1>
          <p className="text-gray-500 mb-6">
            {t('您的订单已支付成功，我们将尽快为您发货。', 'Your order has been paid. We will ship it soon.')}
          </p>
        </>
      ) : status === 'expired' ? (
        <>
          <div className="text-6xl mb-4">⏰</div>
          <h1 className="text-2xl font-bold mb-4">{t('订单已过期', 'Order Expired')}</h1>
          <p className="text-gray-500 mb-6">{t('支付超时，请重新下单。', 'Payment window expired. Please place a new order.')}</p>
        </>
      ) : status === 'syncing' ? (
        <>
          <div className="text-6xl mb-4">⏳</div>
          <h1 className="text-2xl font-bold mb-4">{t('确认支付中...', 'Confirming Payment...')}</h1>
          <p className="text-gray-500 mb-6">{t('正在从 Stripe 同步支付状态，请稍候。', 'Syncing payment status from Stripe...')}</p>
        </>
      ) : (
        <>
          <div className="text-6xl mb-4">📦</div>
          <h1 className="text-2xl font-bold mb-4">{t('订单已创建', 'Order Created')}</h1>
          <p className="text-gray-500 mb-3">
            {t('您的订单已创建成功。支付状态确认中，请稍后查看订单详情。', 'Your order has been created. Payment is being confirmed.')}
          </p>
          {!isAuthenticated && (
            <p className="text-amber-600 text-sm mb-4">
              {t('登录后可查看订单状态。', 'Login to check your order status.')}
            </p>
          )}
        </>
      )}

      <div className="flex justify-center gap-4 flex-wrap">
        {isAuthenticated && (
          <Link
            to={`/orders/${id}`}
            className="px-6 py-2 bg-rose-500 text-white rounded-lg hover:bg-rose-600"
          >
            {t('查看订单', 'View Order')}
          </Link>
        )}
        {!isAuthenticated && (
          <Link
            to={`/login?redirect=/orders/${id}`}
            className="px-6 py-2 bg-rose-500 text-white rounded-lg hover:bg-rose-600"
          >
            {t('登录查看订单', 'Login to View')}
          </Link>
        )}
        <Link
          to="/products"
          className="px-6 py-2 border rounded-lg hover:bg-gray-50"
        >
          {t('继续购物', 'Continue Shopping')}
        </Link>
      </div>

      {status === 'pending' && !orderNo && (
        <p className="mt-4 text-sm text-amber-500">{t('无法确认支付状态，请登录后查看订单。', 'Cannot confirm payment. Please login to check.')}</p>
      )}
    </div>
  )
}
