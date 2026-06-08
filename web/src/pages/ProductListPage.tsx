import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import * as catalogApi from '../api/catalog'
import ProductCard from '../components/shared/ProductCard'

export default function ProductListPage() {
  const locale = useAuthStore((s) => s.locale)
  const [searchParams, setSearchParams] = useSearchParams()
  const page = parseInt(searchParams.get('page') || '1', 10)
  const categoryId = searchParams.get('category')

  const { data: categories } = useQuery({
    queryKey: ['categories'],
    queryFn: catalogApi.getCategories,
  })

  const { data, isLoading } = useQuery({
    queryKey: ['products', page, categoryId],
    queryFn: () => catalogApi.getProducts(page, 12, undefined, categoryId ? parseInt(categoryId) : undefined),
  })

  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)

  const totalPages = data ? Math.ceil(data.total / data.page_size) : 0

  const goPage = (p: number) => {
    const params = new URLSearchParams(searchParams)
    params.set('page', String(p))
    setSearchParams(params)
  }

  return (
    <div className="max-w-7xl mx-auto px-4 py-6">
      <h1 className="text-2xl font-bold mb-6">{t('全部商品', 'All Products')}</h1>

      {/* Category filter */}
      {categories && categories.length > 0 && (
        <div className="flex gap-2 mb-6 flex-wrap">
          <button
            onClick={() => {
              const params = new URLSearchParams(searchParams)
              params.delete('category')
              params.delete('page')
              setSearchParams(params)
            }}
            className={`px-3 py-1 rounded-full text-sm border ${
              !categoryId ? 'bg-rose-50 border-rose-300 text-rose-600' : 'bg-white'
            }`}
          >
            {t('全部', 'All')}
          </button>
          {categories.map((c) => (
            <button
              key={c.id}
              onClick={() => {
                const params = new URLSearchParams(searchParams)
                params.set('category', String(c.id))
                params.delete('page')
                setSearchParams(params)
              }}
              className={`px-3 py-1 rounded-full text-sm border ${
                categoryId === String(c.id) ? 'bg-rose-50 border-rose-300 text-rose-600' : 'bg-white'
              }`}
            >
              {locale === 'zh' ? c.name : c.name_en}
            </button>
          ))}
        </div>
      )}

      {isLoading ? (
        <div className="text-gray-400 text-sm">{t('加载中...', 'Loading...')}</div>
      ) : (data?.products.length || 0) === 0 ? (
        <div className="text-center py-16 text-gray-400">
          {t('暂无商品', 'No products found')}
        </div>
      ) : (
        <>
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
            {(data?.products || []).map((p) => {
              const defaultSku = p.skus?.find((s) => s.is_default) || p.skus?.[0]
              return (
                <ProductCard
                  key={p.id}
                  id={p.id}
                  name={p.name}
                  nameEn={p.name_en}
                  imageUrl={p.image_url || defaultSku?.image_url}
                  price={defaultSku?.price || 0}
                  tags={p.tags}
                  locale={locale}
                />
              )
            })}
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex justify-center gap-2 mt-8">
              <button
                onClick={() => goPage(page - 1)}
                disabled={page <= 1}
                className="px-3 py-1.5 border rounded text-sm disabled:opacity-30"
              >
                {t('上一页', 'Prev')}
              </button>
              {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
                <button
                  key={p}
                  onClick={() => goPage(p)}
                  className={`px-3 py-1.5 rounded text-sm ${
                    p === page ? 'bg-rose-500 text-white' : 'border'
                  }`}
                >
                  {p}
                </button>
              ))}
              <button
                onClick={() => goPage(page + 1)}
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
