import { Link } from 'react-router-dom'
import PriceDisplay from './PriceDisplay'

interface ProductCardProps {
  id: number
  name: string
  nameEn?: string
  imageUrl?: string
  price: number
  tags?: string
  locale: string
}

export default function ProductCard({ id, name, nameEn, imageUrl, price, tags, locale }: ProductCardProps) {
  const displayName = locale === 'en' && nameEn ? nameEn : name
  const tagList = tags ? tags.split(',').filter(Boolean) : []

  return (
    <Link
      to={`/products/${id}`}
      className="block bg-white rounded-xl border hover:shadow-lg transition-shadow overflow-hidden"
    >
      <div className="aspect-square bg-gray-100 flex items-center justify-center">
        {imageUrl ? (
          <img src={imageUrl} alt={displayName} className="w-full h-full object-cover" />
        ) : (
          <span className="text-gray-300 text-4xl">🍿</span>
        )}
      </div>
      <div className="p-3">
        <h3 className="font-medium text-sm truncate">{displayName}</h3>
        <div className="flex items-center gap-2 mt-1">
          <PriceDisplay price={price} className="text-rose-500 font-bold text-sm" />
        </div>
        {tagList.length > 0 && (
          <div className="flex gap-1 mt-2 flex-wrap">
            {tagList.slice(0, 3).map((t) => (
              <span key={t} className="text-xs px-1.5 py-0.5 bg-rose-50 text-rose-500 rounded">
                {t.trim()}
              </span>
            ))}
          </div>
        )}
      </div>
    </Link>
  )
}
