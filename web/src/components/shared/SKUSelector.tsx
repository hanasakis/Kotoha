import type { SKU } from '../../types/catalog'
import PriceDisplay from './PriceDisplay'

interface SKUSelectorProps {
  skus: SKU[]
  selectedId: number | null
  onSelect: (sku: SKU) => void
  locale: string
}

export default function SKUSelector({ skus, selectedId, onSelect, locale }: SKUSelectorProps) {
  if (skus.length <= 1) return null

  return (
    <div className="space-y-2">
      <label className="text-sm font-medium text-gray-700">
        {locale === 'zh' ? '规格' : 'Variant'}
      </label>
      <div className="flex flex-wrap gap-2">
        {skus.map((sku) => (
          <button
            key={sku.id}
            onClick={() => onSelect(sku)}
            disabled={sku.stock === 0}
            className={`px-3 py-1.5 rounded-lg border text-sm transition-colors ${
              selectedId === sku.id
                ? 'border-rose-500 bg-rose-50 text-rose-600'
                : sku.stock === 0
                  ? 'border-gray-200 text-gray-300 cursor-not-allowed'
                  : 'border-gray-200 hover:border-rose-300'
            }`}
          >
            <span>{sku.name}</span>
            <span className="ml-2">
              <PriceDisplay price={sku.price} className="text-xs" />
            </span>
          </button>
        ))}
      </div>
    </div>
  )
}
