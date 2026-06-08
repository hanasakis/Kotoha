import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useParams, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import * as catalogApi from '../api/catalog'
import { useCartStore } from '../store/cartStore'
import PriceDisplay from '../components/shared/PriceDisplay'
import SKUSelector from '../components/shared/SKUSelector'
import type { SKU } from '../types/catalog'

export default function ProductDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const locale = useAuthStore((s) => s.locale)
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const addItem = useCartStore((s) => s.addItem)

  const [selectedSku, setSelectedSku] = useState<SKU | null>(null)
  const [quantity, setQuantity] = useState(1)
  const [adding, setAdding] = useState(false)
  const [added, setAdded] = useState(false)

  const { data: product, isLoading } = useQuery({
    queryKey: ['product', id],
    queryFn: () => catalogApi.getProduct(Number(id)),
    enabled: !!id,
  })

  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)

  if (isLoading) {
    return <div className="max-w-7xl mx-auto px-4 py-8 text-gray-400">{t('加载中...', 'Loading...')}</div>
  }

  if (!product) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-16 text-center text-gray-400">
        {t('商品不存在', 'Product not found')}
      </div>
    )
  }

  const currentSku = selectedSku || product.skus?.find((s) => s.is_default) || product.skus?.[0]
  const inStock = currentSku ? currentSku.stock > 0 : false

  const handleAddToCart = async () => {
    if (!currentSku) return
    if (!isAuthenticated) {
      navigate(`/login?redirect=${encodeURIComponent(`/products/${id}`)}`)
      return
    }
    setAdding(true)
    try {
      await addItem(currentSku.id, quantity)
      setAdded(true)
    } finally {
      setAdding(false)
    }
  }

  const tagList = product.tags ? product.tags.split(',').filter(Boolean) : []

  return (
    <div className="max-w-7xl mx-auto px-4 py-6">
      <div className="grid md:grid-cols-2 gap-8">
        {/* Image */}
        <div className="aspect-square bg-gray-100 rounded-xl flex items-center justify-center">
          {currentSku?.image_url || product.image_url ? (
            <img
              src={currentSku?.image_url || product.image_url}
              alt={locale === 'zh' ? product.name : product.name_en}
              className="w-full h-full object-cover rounded-xl"
            />
          ) : (
            <span className="text-gray-300 text-6xl">🍿</span>
          )}
        </div>

        {/* Info */}
        <div className="space-y-4">
          <h1 className="text-2xl font-bold">
            {locale === 'zh' ? product.name : product.name_en || product.name}
          </h1>

          {tagList.length > 0 && (
            <div className="flex gap-1 flex-wrap">
              {tagList.map((t) => (
                <span key={t} className="text-xs px-2 py-0.5 bg-rose-50 text-rose-500 rounded">
                  {t.trim()}
                </span>
              ))}
            </div>
          )}

          <div className="text-2xl font-bold text-rose-500">
            <PriceDisplay price={currentSku?.price || 0} />
          </div>

          <p className="text-gray-500 text-sm leading-relaxed">
            {locale === 'zh' ? product.description : product.description_en || product.description}
          </p>

          {/* SKU selector */}
          {(product.skus?.length || 0) > 1 && (
            <SKUSelector
              skus={product.skus || []}
              selectedId={currentSku?.id || null}
              onSelect={setSelectedSku}
              locale={locale}
            />
          )}

          {/* Quantity */}
          <div className="space-y-2">
            <label className="text-sm font-medium text-gray-700">
              {t('数量', 'Quantity')}
            </label>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setQuantity(Math.max(1, quantity - 1))}
                className="w-8 h-8 border rounded-lg text-lg leading-none"
              >
                -
              </button>
              <span className="w-10 text-center">{quantity}</span>
              <button
                onClick={() => setQuantity(Math.min(currentSku?.stock || 99, quantity + 1))}
                className="w-8 h-8 border rounded-lg text-lg leading-none"
              >
                +
              </button>
            </div>
          </div>

          {/* Add to cart */}
          <button
            onClick={handleAddToCart}
            disabled={!inStock || adding || added}
            className={`w-full py-3 rounded-lg font-medium text-white transition-colors ${
              added
                ? 'bg-green-500'
                : inStock
                  ? 'bg-rose-500 hover:bg-rose-600'
                  : 'bg-gray-300 cursor-not-allowed'
            }`}
          >
            {!inStock
              ? t('已售罄', 'Out of Stock')
              : added
                ? t('已加入购物车 ✓', 'Added to Cart ✓')
                : adding
                  ? t('加入中...', 'Adding...')
                  : t('加入购物车', 'Add to Cart')}
          </button>

          {/* Product meta */}
          <div className="border-t pt-4 text-sm text-gray-400 space-y-1">
            {product.weight_gram > 0 && (
              <p>{t('净重', 'Net weight')}: {product.weight_gram}g</p>
            )}
            {product.shelf_days > 0 && (
              <p>{t('保质期', 'Shelf life')}: {product.shelf_days}{t('天', ' days')}</p>
            )}
            {product.ingredients && (
              <p>{t('配料', 'Ingredients')}: {product.ingredients}</p>
            )}
            {product.allergens && (
              <p>{t('过敏原', 'Allergens')}: {product.allergens}</p>
            )}
            {product.taste && (
              <p>{t('口味', 'Taste')}: {product.taste}</p>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
