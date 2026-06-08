import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useNavigate, useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useAuthStore } from '../store/authStore'
import * as userApi from '../api/user'

const schema = z.object({
  name: z.string().min(1, '请输入收货人姓名'),
  phone: z.string().min(1, '请输入手机号'),
  province: z.string().min(1, '请输入省份'),
  city: z.string().min(1, '请输入城市'),
  district: z.string().min(1, '请输入区县'),
  detail: z.string().min(1, '请输入详细地址'),
  is_default: z.boolean().optional(),
})

type FormData = z.infer<typeof schema>

export default function AddressFormPage() {
  const { id } = useParams<{ id: string }>()
  const isEdit = !!id
  const locale = useAuthStore((s) => s.locale)
  const navigate = useNavigate()
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)

  const { data: addresses } = useQuery({
    queryKey: ['addresses'],
    queryFn: userApi.listAddresses,
    enabled: isEdit,
  })

  const existing = isEdit ? (addresses || []).find((a) => a.id === Number(id)) : null

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
    values: existing
      ? {
          name: existing.name,
          phone: existing.phone,
          province: existing.province,
          city: existing.city,
          district: existing.district,
          detail: existing.detail,
          is_default: existing.is_default,
        }
      : undefined,
  })

  const onSubmit = async (data: FormData) => {
    setError('')
    setLoading(true)
    try {
      if (isEdit) {
        await userApi.updateAddress(Number(id), data)
      } else {
        await userApi.createAddress(data)
      }
      navigate('/profile/addresses')
    } catch (err: unknown) {
      const e = err as { code?: string; message?: string }
      setError(e.message || t('保存失败', 'Save failed'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="max-w-md mx-auto px-4 py-6">
      <h1 className="text-2xl font-bold mb-6">
        {isEdit ? t('编辑地址', 'Edit Address') : t('添加地址', 'Add Address')}
      </h1>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        {error && (
          <div className="bg-red-50 text-red-600 text-sm px-4 py-3 rounded-lg">{error}</div>
        )}

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            {t('收货人', 'Name')}
          </label>
          <input
            type="text"
            {...register('name')}
            className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-rose-400"
          />
          {errors.name && <p className="text-red-500 text-xs mt-1">{errors.name.message}</p>}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            {t('手机号', 'Phone')}
          </label>
          <input
            type="text"
            {...register('phone')}
            className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-rose-400"
          />
          {errors.phone && <p className="text-red-500 text-xs mt-1">{errors.phone.message}</p>}
        </div>

        <div className="grid grid-cols-3 gap-2">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              {t('省份', 'Province')}
            </label>
            <input
              type="text"
              {...register('province')}
              className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-rose-400"
            />
            {errors.province && <p className="text-red-500 text-xs mt-1">{errors.province.message}</p>}
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              {t('城市', 'City')}
            </label>
            <input
              type="text"
              {...register('city')}
              className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-rose-400"
            />
            {errors.city && <p className="text-red-500 text-xs mt-1">{errors.city.message}</p>}
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              {t('区县', 'District')}
            </label>
            <input
              type="text"
              {...register('district')}
              className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-rose-400"
            />
            {errors.district && <p className="text-red-500 text-xs mt-1">{errors.district.message}</p>}
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            {t('详细地址', 'Detail')}
          </label>
          <input
            type="text"
            {...register('detail')}
            className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-rose-400"
          />
          {errors.detail && <p className="text-red-500 text-xs mt-1">{errors.detail.message}</p>}
        </div>

        <label className="flex items-center gap-2 text-sm">
          <input type="checkbox" {...register('is_default')} className="accent-rose-500" />
          {t('设为默认地址', 'Set as default')}
        </label>

        <button
          type="submit"
          disabled={loading}
          className="w-full bg-rose-500 text-white py-2 rounded-lg hover:bg-rose-600 disabled:opacity-50"
        >
          {loading ? t('保存中...', 'Saving...') : t('保存', 'Save')}
        </button>
      </form>
    </div>
  )
}
