import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import * as userApi from '../api/user'

export default function AddressPage() {
  const locale = useAuthStore((s) => s.locale)
  const queryClient = useQueryClient()
  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)

  const { data: addresses, isLoading } = useQuery({
    queryKey: ['addresses'],
    queryFn: userApi.listAddresses,
  })

  const handleDelete = async (id: number) => {
    if (!confirm(t('确定删除此地址？', 'Delete this address?'))) return
    await userApi.deleteAddress(id)
    queryClient.invalidateQueries({ queryKey: ['addresses'] })
  }

  return (
    <div className="max-w-3xl mx-auto px-4 py-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">{t('收货地址', 'Addresses')}</h1>
        <Link
          to="/profile/addresses/new"
          className="text-sm bg-rose-500 text-white px-3 py-1.5 rounded-lg hover:bg-rose-600"
        >
          + {t('添加', 'Add')}
        </Link>
      </div>

      {isLoading ? (
        <div className="text-gray-400 text-sm">{t('加载中...', 'Loading...')}</div>
      ) : (addresses?.length || 0) === 0 ? (
        <div className="text-center py-16 text-gray-400">
          {t('暂无地址', 'No addresses yet')}
        </div>
      ) : (
        <div className="space-y-3">
          {(addresses || []).map((a) => (
            <div key={a.id} className="p-4 bg-white border rounded-xl flex justify-between items-start">
              <div>
                <p className="font-medium">
                  {a.name} <span className="text-sm text-gray-400">{a.phone}</span>
                  {a.is_default && (
                    <span className="ml-2 text-xs text-rose-500">{t('默认', 'Default')}</span>
                  )}
                </p>
                <p className="text-sm text-gray-500 mt-1">
                  {a.province} {a.city} {a.district} {a.detail}
                </p>
              </div>
              <div className="flex gap-2 text-sm shrink-0">
                <Link
                  to={`/profile/addresses/${a.id}/edit`}
                  className="text-gray-400 hover:text-gray-600"
                >
                  {t('编辑', 'Edit')}
                </Link>
                <button
                  onClick={() => handleDelete(a.id)}
                  className="text-gray-400 hover:text-red-500"
                >
                  {t('删除', 'Delete')}
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
