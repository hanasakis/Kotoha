import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useAuthStore } from '../../store/authStore'
import * as catalogApi from '../../api/catalog'
import * as adminApi from '../../api/admin'
import type { Product } from '../../types/catalog'

function ProductForm({
  product,
  onClose,
}: {
  product?: Product | null
  onClose: () => void
}) {
  const queryClient = useQueryClient()
  const isEdit = !!product
  const locale = useAuthStore((s) => s.locale)
  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)

  const [form, setForm] = useState({
    name: product?.name || '',
    name_en: product?.name_en || '',
    description: product?.description || '',
    description_en: product?.description_en || '',
    image_url: product?.image_url || '',
    category_id: product?.category_id || 1,
    tags: product?.tags || '',
    taste: product?.taste || '',
    allergens: product?.allergens || '',
    ingredients: product?.ingredients || '',
    weight_gram: product?.weight_gram || 100,
    shelf_days: product?.shelf_days || 180,
    is_active: product?.is_active ?? true,
  })

  const createMutation = useMutation({
    mutationFn: adminApi.createProduct,
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['products'] }); onClose() },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<Product> }) => adminApi.updateProduct(id, data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['products'] }); onClose() },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    const payload = { ...form }
    if (isEdit && product) {
      updateMutation.mutate({ id: product.id, data: payload })
    } else {
      createMutation.mutate(payload as any)
    }
  }

  const isPending = createMutation.isPending || updateMutation.isPending

  return (
    <div className="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-lg max-h-[90vh] overflow-y-auto">
        <div className="px-6 py-4 border-b flex justify-between items-center">
          <h3 className="font-semibold">
            {isEdit ? t('编辑商品', 'Edit Product') : t('新增商品', 'New Product')}
          </h3>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600 text-lg">&times;</button>
        </div>
        <form onSubmit={handleSubmit} className="px-6 py-4 space-y-3">
          <div className="grid grid-cols-2 gap-3">
            <label className="block">
              <span className="text-xs text-gray-500">{t('商品名', 'Name')}</span>
              <input className="w-full border rounded px-2 py-1.5 text-sm mt-0.5" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
            </label>
            <label className="block">
              <span className="text-xs text-gray-500">{t('英文名', 'Name EN')}</span>
              <input className="w-full border rounded px-2 py-1.5 text-sm mt-0.5" value={form.name_en} onChange={(e) => setForm({ ...form, name_en: e.target.value })} />
            </label>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <label className="block">
              <span className="text-xs text-gray-500">{t('分类ID', 'Category ID')}</span>
              <input type="number" className="w-full border rounded px-2 py-1.5 text-sm mt-0.5" value={form.category_id} onChange={(e) => setForm({ ...form, category_id: +e.target.value })} />
            </label>
            <label className="block">
              <span className="text-xs text-gray-500">{t('图片URL', 'Image URL')}</span>
              <input className="w-full border rounded px-2 py-1.5 text-sm mt-0.5" value={form.image_url} onChange={(e) => setForm({ ...form, image_url: e.target.value })} />
            </label>
          </div>
          <label className="block">
            <span className="text-xs text-gray-500">{t('描述', 'Description')}</span>
            <textarea rows={2} className="w-full border rounded px-2 py-1.5 text-sm mt-0.5" value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
          </label>
          <label className="block">
            <span className="text-xs text-gray-500">{t('英文描述', 'Description EN')}</span>
            <textarea rows={2} className="w-full border rounded px-2 py-1.5 text-sm mt-0.5" value={form.description_en} onChange={(e) => setForm({ ...form, description_en: e.target.value })} />
          </label>
          <div className="grid grid-cols-2 gap-3">
            <label className="block">
              <span className="text-xs text-gray-500">{t('口味', 'Taste')}</span>
              <input className="w-full border rounded px-2 py-1.5 text-sm mt-0.5" value={form.taste} onChange={(e) => setForm({ ...form, taste: e.target.value })} />
            </label>
            <label className="block">
              <span className="text-xs text-gray-500">{t('标签', 'Tags')}</span>
              <input className="w-full border rounded px-2 py-1.5 text-sm mt-0.5" value={form.tags} onChange={(e) => setForm({ ...form, tags: e.target.value })} />
            </label>
          </div>
          <div className="grid grid-cols-3 gap-3">
            <label className="block">
              <span className="text-xs text-gray-500">{t('重量(g)', 'Weight(g)')}</span>
              <input type="number" className="w-full border rounded px-2 py-1.5 text-sm mt-0.5" value={form.weight_gram} onChange={(e) => setForm({ ...form, weight_gram: +e.target.value })} />
            </label>
            <label className="block">
              <span className="text-xs text-gray-500">{t('保质期(天)', 'Shelf(d)')}</span>
              <input type="number" className="w-full border rounded px-2 py-1.5 text-sm mt-0.5" value={form.shelf_days} onChange={(e) => setForm({ ...form, shelf_days: +e.target.value })} />
            </label>
            <label className="flex items-center gap-2 pt-5">
              <input type="checkbox" checked={form.is_active} onChange={(e) => setForm({ ...form, is_active: e.target.checked })} />
              <span className="text-xs text-gray-500">{t('上架', 'Active')}</span>
            </label>
          </div>
          <div className="flex justify-end gap-3 pt-2">
            <button type="button" onClick={onClose} className="px-4 py-1.5 text-sm text-gray-500 hover:text-gray-700">
              {t('取消', 'Cancel')}
            </button>
            <button type="submit" disabled={isPending} className="px-4 py-1.5 text-sm bg-rose-500 text-white rounded-lg hover:bg-rose-600 disabled:opacity-50">
              {isPending ? t('保存中...', 'Saving...') : t('保存', 'Save')}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default function AdminProductsPage() {
  const locale = useAuthStore((s) => s.locale)
  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [editingProduct, setEditingProduct] = useState<Product | null | undefined>(undefined)

  const { data, isLoading } = useQuery({
    queryKey: ['products', page, 20],
    queryFn: () => catalogApi.getProducts(page, 20),
  })

  const deleteMutation = useMutation({
    mutationFn: adminApi.deleteProduct,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['products'] }),
  })

  const reindexMutation = useMutation({
    mutationFn: adminApi.reindexSearch,
    onSuccess: () => alert(t('索引重建成功', 'Reindex succeeded')),
  })

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-xl font-bold text-gray-900">{t('商品管理', 'Products')}</h1>
        <div className="flex gap-3">
          <button
            onClick={() => reindexMutation.mutate()}
            disabled={reindexMutation.isPending}
            className="px-4 py-2 text-sm border rounded-lg hover:bg-gray-100 disabled:opacity-50"
          >
            {reindexMutation.isPending ? t('索引中...', 'Indexing...') : t('重建搜索索引', 'Rebuild Search Index')}
          </button>
          <button
            onClick={() => setEditingProduct(null)}
            className="px-4 py-2 text-sm bg-rose-500 text-white rounded-lg hover:bg-rose-600"
          >
            {t('新增商品', 'Add Product')}
          </button>
        </div>
      </div>

      <div className="bg-white border rounded-xl overflow-hidden">
        {isLoading ? (
          <div className="p-8 text-center text-gray-400">{t('加载中...', 'Loading...')}</div>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-gray-500 border-b bg-gray-50">
                <th className="py-3 px-4 font-medium">ID</th>
                <th className="py-3 px-4 font-medium">{t('名称', 'Name')}</th>
                <th className="py-3 px-4 font-medium">{t('分类', 'Cat.')}</th>
                <th className="py-3 px-4 font-medium">{t('状态', 'Status')}</th>
                <th className="py-3 px-4 font-medium">{t('点击', 'Clicks')}</th>
                <th className="py-3 px-4 font-medium">{t('购买', 'Buys')}</th>
                <th className="py-3 px-4 font-medium">{t('操作', 'Actions')}</th>
              </tr>
            </thead>
            <tbody>
              {(data?.products || []).map((p) => (
                <tr key={p.id} className="border-t hover:bg-gray-50">
                  <td className="py-2.5 px-4 text-gray-500">{p.id}</td>
                  <td className="py-2.5 px-4">
                    <span className="font-medium">{p.name}</span>
                    {p.name_en && <span className="text-gray-400 ml-1.5 text-xs">({p.name_en})</span>}
                  </td>
                  <td className="py-2.5 px-4 text-gray-500">{p.category_id}</td>
                  <td className="py-2.5 px-4">
                    <span className={`text-xs px-2 py-0.5 rounded-full ${p.is_active ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'}`}>
                      {p.is_active ? t('上架', 'Active') : t('下架', 'Off')}
                    </span>
                  </td>
                  <td className="py-2.5 px-4 text-gray-500">{p.click_count}</td>
                  <td className="py-2.5 px-4 text-gray-500">{p.buy_count}</td>
                  <td className="py-2.5 px-4">
                    <button onClick={() => setEditingProduct(p)} className="text-rose-500 hover:underline mr-3 text-xs">
                      {t('编辑', 'Edit')}
                    </button>
                    <button
                      onClick={() => { if (confirm(t('确定删除？', 'Confirm delete?'))) deleteMutation.mutate(p.id) }}
                      className="text-red-400 hover:underline text-xs"
                    >
                      {t('删除', 'Del')}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}

        {/* Pagination */}
        {data && data.total > 20 && (
          <div className="px-4 py-3 border-t flex justify-between items-center text-sm text-gray-500">
            <span>{t(`共 ${data.total} 件商品`, `Total ${data.total} products`)}</span>
            <div className="flex gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page <= 1}
                className="px-3 py-1 border rounded hover:bg-gray-100 disabled:opacity-30"
              >
                {t('上一页', 'Prev')}
              </button>
              <span className="px-3 py-1">{page}</span>
              <button
                onClick={() => setPage((p) => p + 1)}
                disabled={page * 20 >= data.total}
                className="px-3 py-1 border rounded hover:bg-gray-100 disabled:opacity-30"
              >
                {t('下一页', 'Next')}
              </button>
            </div>
          </div>
        )}
      </div>

      {editingProduct !== undefined && (
        <ProductForm product={editingProduct} onClose={() => setEditingProduct(undefined)} />
      )}
    </div>
  )
}
