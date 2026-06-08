import { Link } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import { useCartStore } from '../store/cartStore'
import PriceDisplay from '../components/shared/PriceDisplay'

export default function CartPage() {
  const locale = useAuthStore((s) => s.locale)
  const { items, loading, updateQty, removeItem } = useCartStore()
  const total = items.reduce((sum, i) => sum + i.price * i.quantity, 0)

  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)

  if (loading) {
    return <div className="max-w-3xl mx-auto px-4 py-8 text-gray-400">{t('加载中...', 'Loading...')}</div>
  }

  return (
    <div className="max-w-3xl mx-auto px-4 py-6">
      <h1 className="text-2xl font-bold mb-6">{t('购物车', 'Cart')}</h1>

      {items.length === 0 ? (
        <div className="text-center py-16 text-gray-400">
          <p className="mb-4">{t('购物车是空的', 'Your cart is empty')}</p>
          <Link to="/products" className="text-rose-500 hover:underline">
            {t('去逛逛', 'Browse Products')} →
          </Link>
        </div>
      ) : (
        <>
          <div className="space-y-3">
            {items.map((item) => (
              <div key={item.sku_id} className="flex gap-4 p-4 bg-white border rounded-xl">
                <div className="w-20 h-20 bg-gray-100 rounded-lg shrink-0 flex items-center justify-center">
                  {item.image_url ? (
                    <img src={item.image_url} alt={item.product_name} className="w-full h-full object-cover rounded-lg" />
                  ) : (
                    <span className="text-gray-300 text-2xl">🍿</span>
                  )}
                </div>
                <div className="flex-1 min-w-0">
                  <Link to={`/products/${item.product_id}`} className="font-medium text-sm hover:text-rose-500">
                    {item.product_name}
                  </Link>
                  <p className="text-xs text-gray-400">{item.sku_name}</p>
                  <div className="flex items-center justify-between mt-2">
                    <PriceDisplay price={item.price} className="text-rose-500 font-bold" />
                    <div className="flex items-center gap-1">
                      <button
                        onClick={() => updateQty(item.sku_id, Math.max(1, item.quantity - 1))}
                        className="w-7 h-7 border rounded text-sm leading-none"
                      >
                        -
                      </button>
                      <span className="w-8 text-center text-sm">{item.quantity}</span>
                      <button
                        onClick={() => updateQty(item.sku_id, Math.min(item.stock, item.quantity + 1))}
                        className="w-7 h-7 border rounded text-sm leading-none"
                      >
                        +
                      </button>
                      <button
                        onClick={() => removeItem(item.sku_id)}
                        className="ml-3 text-xs text-gray-400 hover:text-red-500"
                      >
                        {t('删除', 'Delete')}
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>

          <div className="mt-6 p-4 bg-white border rounded-xl">
            <div className="flex justify-between items-center text-lg font-bold">
              <span>{t('合计', 'Total')}</span>
              <PriceDisplay price={total} className="text-rose-500" />
            </div>
            <Link
              to="/checkout"
              className="block text-center mt-4 w-full bg-rose-500 text-white py-3 rounded-lg hover:bg-rose-600 font-medium"
            >
              {t('去结算', 'Checkout')}
            </Link>
          </div>
        </>
      )}
    </div>
  )
}
