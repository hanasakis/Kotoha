import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import * as catalogApi from '../api/catalog'
import ProductCard from '../components/shared/ProductCard'

export default function HomePage() {
  const locale = useAuthStore((s) => s.locale)

  const { data: categories, isLoading: catLoading } = useQuery({
    queryKey: ['categories'],
    queryFn: catalogApi.getCategories,
  })

  const { data: productsData, isLoading: prodLoading } = useQuery({
    queryKey: ['products', 1],
    queryFn: () => catalogApi.getProducts(1, 8),
  })

  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)
  const catName = (c: { name: string; name_en: string }) => (locale === 'zh' ? c.name : c.name_en)

  return (
    <div className="max-w-7xl mx-auto px-4 py-6 space-y-10">
      {/* Hero */}
      <section className="text-center py-12 bg-gradient-to-r from-rose-50 to-orange-50 rounded-2xl">
        <h1 className="text-3xl font-bold text-gray-900 mb-3">
          {t('发现你的美味零食', 'Discover Your Snacks')}
        </h1>
        <p className="text-gray-500 mb-6">
          {t('精选全球零食，直达你的舌尖', 'Curated snacks from around the world')}
        </p>
        <Link
          to="/products"
          className="inline-block bg-rose-500 text-white px-6 py-2.5 rounded-lg hover:bg-rose-600 font-medium"
        >
          {t('浏览全部商品', 'Browse All')}
        </Link>
      </section>

      {/* Categories */}
      <section>
        <h2 className="text-lg font-bold mb-4">{t('商品分类', 'Categories')}</h2>
        {catLoading ? (
          <div className="text-gray-400 text-sm">{t('加载中...', 'Loading...')}</div>
        ) : (
          <div className="flex gap-3 overflow-x-auto pb-2">
            {(categories || []).map((c) => (
              <Link
                key={c.id}
                to={`/products?category=${c.id}`}
                className="shrink-0 px-4 py-2 bg-white border rounded-full text-sm hover:border-rose-300 hover:text-rose-500 transition-colors"
              >
                {catName(c)}
              </Link>
            ))}
          </div>
        )}
      </section>

      {/* Featured Products */}
      <section>
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-lg font-bold">{t('推荐商品', 'Featured')}</h2>
          <Link to="/products" className="text-sm text-rose-500 hover:underline">
            {t('查看全部', 'View All')} →
          </Link>
        </div>
        {prodLoading ? (
          <div className="text-gray-400 text-sm">{t('加载中...', 'Loading...')}</div>
        ) : (
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {(productsData?.products || []).map((p) => {
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
        )}
      </section>
    </div>
  )
}
