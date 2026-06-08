import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Link, useSearchParams, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import * as authApi from '../api/auth'

const schema = z.object({
  new_password: z.string().min(6, '密码至少6位'),
})

type FormData = z.infer<typeof schema>

export default function ResetPasswordPage() {
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(false)
  const [loading, setLoading] = useState(false)
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const locale = useAuthStore((s) => s.locale)

  const token = searchParams.get('token') || ''

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
  })

  const onSubmit = async (data: FormData) => {
    if (!token) {
      setError(locale === 'zh' ? '缺少重置令牌' : 'Missing reset token')
      return
    }
    setError('')
    setLoading(true)
    try {
      await authApi.resetPassword({ token, new_password: data.new_password })
      setSuccess(true)
    } catch (err: unknown) {
      const e = err as { code?: string; message?: string }
      setError(e.message || (locale === 'zh' ? '重置失败，请重试' : 'Reset failed, please retry'))
    } finally {
      setLoading(false)
    }
  }

  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)

  if (success) {
    return (
      <div className="max-w-md mx-auto mt-16 px-4 text-center">
        <h1 className="text-2xl font-bold mb-4">
          {t('密码已重置', 'Password Reset')}
        </h1>
        <p className="text-gray-500 mb-6">
          {t('您的密码已成功重置，请使用新密码登录。', 'Your password has been reset. Please log in with your new password.')}
        </p>
        <button
          onClick={() => navigate('/login')}
          className="bg-rose-500 text-white px-6 py-2 rounded-lg hover:bg-rose-600"
        >
          {t('去登录', 'Go to Login')}
        </button>
      </div>
    )
  }

  if (!token) {
    return (
      <div className="max-w-md mx-auto mt-16 px-4 text-center">
        <h1 className="text-2xl font-bold mb-4">
          {t('无效的重置链接', 'Invalid Reset Link')}
        </h1>
        <p className="text-gray-500 mb-6">
          {t('此密码重置链接无效或已过期，请重新申请密码重置。', 'This password reset link is invalid or expired. Please request a new one.')}
        </p>
        <Link to="/forgot-password" className="text-rose-500 hover:underline">
          {t('重新申请', 'Request Again')}
        </Link>
      </div>
    )
  }

  return (
    <div className="max-w-md mx-auto mt-16 px-4">
      <h1 className="text-2xl font-bold text-center mb-8">
        {t('重置密码', 'Reset Password')}
      </h1>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        {error && (
          <div className="bg-red-50 text-red-600 text-sm px-4 py-3 rounded-lg">{error}</div>
        )}

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            {t('新密码', 'New Password')}
          </label>
          <input
            type="password"
            {...register('new_password')}
            className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-rose-400"
            placeholder={t('请输入新密码（至少6位）', 'Enter new password (min 6 chars)')}
          />
          {errors.new_password && (
            <p className="text-red-500 text-xs mt-1">{errors.new_password.message}</p>
          )}
        </div>

        <button
          type="submit"
          disabled={loading}
          className="w-full bg-rose-500 text-white py-2 rounded-lg hover:bg-rose-600 disabled:opacity-50"
        >
          {loading
            ? t('重置中...', 'Resetting...')
            : t('重置密码', 'Reset Password')}
        </button>

        <p className="text-center text-sm text-gray-500">
          <Link to="/login" className="text-rose-500 hover:underline">
            {t('返回登录', 'Back to Login')}
          </Link>
        </p>
      </form>
    </div>
  )
}
