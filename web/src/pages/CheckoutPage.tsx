import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import { useCartStore } from '../store/cartStore'
import * as userApi from '../api/user'
import * as orderApi from '../api/order'
import * as paymentApi from '../api/payment'
import PriceDisplay from '../components/shared/PriceDisplay'
import type { Address } from '../types/user'

export default function CheckoutPage() {
  const locale = useAuthStore((s) => s.locale)
  const { items, fetchCart } = useCartStore()
  const navigate = useNavigate()

  const [addresses, setAddresses] = useState<Address[]>([])
  const [selectedAddr, setSelectedAddr] = useState<number | null>(null)
  const [note, setNote] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [step, setStep] = useState<'loading' | 'ready' | 'submitting'>('loading')

  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)
  const total = items.reduce((sum, i) => sum + i.price * i.quantity, 0)

  useEffect(() => {
    fetchCart().then(() => {
      userApi.listAddresses().then((addrs) => {
        setAddresses(addrs)
        const def = addrs.find((a) => a.is_default) || addrs[0]
        if (def) setSelectedAddr(def.id)
        setStep('ready')
      })
    })
  }, [fetchCart])

  const handleSubmit = async () => {
    if (!selectedAddr) {
      setError(t('请选择收货地址', 'Please select an address'))
      return
    }
    setError('')
    setStep('submitting')
    try {
      const order = await orderApi.createOrder({ address_id: selectedAddr, note: note || undefined })
      const base = window.location.origin
      const result = await paymentApi.createCheckout(order.id, {
        success_url: `${base}/orders/${order.id}/success?order_no=${encodeURIComponent(order.order_no)}`,
        cancel_url: `${base}/orders/${order.id}/cancel`,
      })
      // Redirect to Stripe checkout
      window.location.href = result.checkout_url
    } catch (err: unknown) {
      const e = err as { code?: string; message?: string }
      setError(e.message || t('下单失败，请重试', 'Order failed, please retry'))
      setStep('ready')
    }
  }

  if (step === 'loading') {
    return <div className="max-w-3xl mx-auto px-4 py-8 text-gray-400">{t('加载中...', 'Loading...')}</div>
  }

  if (items.length === 0) {
    return (
      <div className="max-w-3xl mx-auto px-4 py-16 text-center text-gray-400">
        <p>{t('购物车是空的', 'Your cart is empty')}</p>
      </div>
    )
  }

  return (
    <div className="max-w-3xl mx-auto px-4 py-6">
      <h1 className="text-2xl font-bold mb-6">{t('确认订单', 'Checkout')}</h1>

      {error && (
        <div className="bg-red-50 text-red-600 text-sm px-4 py-3 rounded-lg mb-4">{error}</div>
      )}

      {/* Address */}
      <section className="mb-6">
        <h2 className="font-bold mb-3">{t('收货地址', 'Shipping Address')}</h2>
        {addresses.length === 0 ? (
          <div className="text-gray-400 text-sm">
            <p className="mb-2">{t('暂无地址', 'No addresses yet')}</p>
            <button
              onClick={() => navigate('/profile/addresses/new')}
              className="text-rose-500 hover:underline"
            >
              {t('添加地址', 'Add Address')}
            </button>
          </div>
        ) : (
          <div className="space-y-2">
            {addresses.map((a) => (
              <label
                key={a.id}
                className={`flex items-start gap-3 p-3 border rounded-lg cursor-pointer ${
                  selectedAddr === a.id ? 'border-rose-500 bg-rose-50' : ''
                }`}
              >
                <input
                  type="radio"
                  name="address"
                  checked={selectedAddr === a.id}
                  onChange={() => setSelectedAddr(a.id)}
                  className="mt-1 accent-rose-500"
                />
                <div className="text-sm">
                  <p className="font-medium">{a.name} {a.phone}</p>
                  <p className="text-gray-500">{a.province} {a.city} {a.district} {a.detail}</p>
                  {a.is_default && <span className="text-xs text-rose-500">{t('默认', 'Default')}</span>}
                </div>
              </label>
            ))}
            <button
              onClick={() => navigate('/profile/addresses/new')}
              className="text-sm text-rose-500 hover:underline"
            >
              + {t('添加新地址', 'Add New Address')}
            </button>
          </div>
        )}
      </section>

      {/* Order summary */}
      <section className="mb-6">
        <h2 className="font-bold mb-3">{t('商品明细', 'Order Summary')}</h2>
        <div className="border rounded-xl divide-y">
          {items.map((item) => (
            <div key={item.sku_id} className="flex justify-between p-3 text-sm">
              <div>
                <span>{item.product_name}</span>
                <span className="text-gray-400 ml-2">{item.sku_name} x{item.quantity}</span>
              </div>
              <PriceDisplay price={item.price * item.quantity} />
            </div>
          ))}
        </div>

        <div className="mt-3">
          <label className="text-sm font-medium text-gray-700">{t('订单备注', 'Note')}</label>
          <input
            type="text"
            value={note}
            onChange={(e) => setNote(e.target.value)}
            className="w-full mt-1 px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-rose-400"
            placeholder={t('选填', 'Optional')}
          />
        </div>
      </section>

      {/* Total + Submit */}
      <div className="p-4 bg-white border rounded-xl">
        <div className="flex justify-between items-center text-lg font-bold mb-4">
          <span>{t('合计', 'Total')}</span>
          <PriceDisplay price={total} className="text-rose-500" />
        </div>
        <button
          onClick={handleSubmit}
          disabled={step === 'submitting' || !selectedAddr}
          className="w-full bg-rose-500 text-white py-3 rounded-lg hover:bg-rose-600 disabled:opacity-50 font-medium"
        >
          {step === 'submitting' ? t('提交中...', 'Submitting...') : t('提交订单', 'Place Order')}
        </button>
      </div>
    </div>
  )
}
