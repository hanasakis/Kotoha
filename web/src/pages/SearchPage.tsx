import { useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import * as catalogApi from '../api/catalog'
import ProductCard from '../components/shared/ProductCard'

export default function SearchPage() {
  const locale = useAuthStore((s) => s.locale)
  const [searchParams] = useSearchParams()
  const query = searchParams.get('q') || ''

  const { data, isLoading } = useQuery({
    queryKey: ['search', query],
    queryFn: () => catalogApi.getProducts(1, 50, query),
    enabled: query.length > 0,
  })

  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)

  return (
    <div className="max-w-7xl mx-auto px-4 py-6">
      <h1 className="text-2xl font-bold mb-2">
        {t('搜索结果', 'Search Results')}
      </h1>
      <p className="text-gray-500 mb-6">
        "{query}" {(data?.products?.length || 0)}{t('个结果', ' results')}
      </p>

      {isLoading ? (
        <div className="text-gray-400 text-sm">{t('加载中...', 'Loading...')}</div>
      ) : !query ? (
        <div className="text-center py-16 text-gray-400">
          {t('请输入搜索关键词', 'Please enter a search term')}
        </div>
      ) : (data?.products.length || 0) === 0 ? (
        <div className="text-center py-16 text-gray-400">
          {t('未找到相关商品', 'No products found')}
        </div>
      ) : (
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
      )}
    </div>
  )
}
