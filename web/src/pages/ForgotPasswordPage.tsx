import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Link } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import * as authApi from '../api/auth'

const normalizeFullWidth = (v: unknown) =>
  typeof v === 'string'
    ? v.replace(/[！-～]/g, (ch) =>
        String.fromCharCode(ch.charCodeAt(0) - 0xFEE0)
      )
    : v

const schema = z.object({
  email: z.preprocess(
    normalizeFullWidth,
    z.string().min(1, '请输入邮箱').email('邮箱格式不正确'),
  ),
})

type FormData = z.infer<typeof schema>

export default function ForgotPasswordPage() {
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(false)
  const [loading, setLoading] = useState(false)
  const locale = useAuthStore((s) => s.locale)

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormData>({
    resolver: zodResolver(schema),
  })

  const onSubmit = async (data: FormData) => {
    setError('')
    setLoading(true)
    try {
      await authApi.forgotPassword(data)
      setSuccess(true)
    } catch (err: unknown) {
      const e = err as { code?: string; message?: string }
      setError(e.message || (locale === 'zh' ? '发送失败，请重试' : 'Failed to send, please retry'))
    } finally {
      setLoading(false)
    }
  }

  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)

  if (success) {
    return (
      <div className="max-w-md mx-auto mt-16 px-4 text-center">
        <h1 className="text-2xl font-bold mb-4">
          {t('邮件已发送', 'Email Sent')}
        </h1>
        <p className="text-gray-500 mb-6">
          {t('请检查您的邮箱，点击邮件中的链接重置密码。', 'Please check your email and click the link to reset your password.')}
        </p>
        <Link to="/login" className="text-rose-500 hover:underline">
          {t('返回登录', 'Back to Login')}
        </Link>
      </div>
    )
  }

  return (
    <div className="max-w-md mx-auto mt-16 px-4">
      <h1 className="text-2xl font-bold text-center mb-8">
        {t('忘记密码', 'Forgot Password')}
      </h1>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        {error && (
          <div className="bg-red-50 text-red-600 text-sm px-4 py-3 rounded-lg">{error}</div>
        )}

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            {t('邮箱', 'Email')}
          </label>
          <input
            type="email"
            {...register('email')}
            className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-rose-400"
            placeholder={t('请输入注册邮箱', 'Enter your registered email')}
          />
          {errors.email && (
            <p className="text-red-500 text-xs mt-1">{errors.email.message}</p>
          )}
        </div>

        <button
          type="submit"
          disabled={loading}
          className="w-full bg-rose-500 text-white py-2 rounded-lg hover:bg-rose-600 disabled:opacity-50"
        >
          {loading
            ? t('发送中...', 'Sending...')
            : t('发送重置邮件', 'Send Reset Email')}
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
