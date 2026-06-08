import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Link, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'

const normalizeFullWidth = (v: unknown) =>
  typeof v === 'string'
    ? v.replace(/[！-～]/g, (ch) =>
        String.fromCharCode(ch.charCodeAt(0) - 0xFEE0)
      )
    : v

const schema = z.object({
  nickname: z.string().min(1, '请输入昵称').max(50, '昵称最多50个字符'),
  email: z.preprocess(
    normalizeFullWidth,
    z.string().min(1, '请输入邮箱').email('邮箱格式不正确'),
  ),
  password: z.string().min(6, '密码至少6位'),
})

type FormData = z.infer<typeof schema>

export default function RegisterPage() {
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const reg = useAuthStore((s) => s.register)
  const navigate = useNavigate()
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
      await reg(data.email, data.password, data.nickname)
      navigate('/', { replace: true })
    } catch (err: unknown) {
      const e = err as { code?: string; message?: string }
      if (e.code === 'auth.email_exists') {
        setError(locale === 'zh' ? '该邮箱已注册' : 'Email already registered')
      } else {
        setError(e.message || (locale === 'zh' ? '注册失败' : 'Registration failed'))
      }
    } finally {
      setLoading(false)
    }
  }

  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)

  return (
    <div className="max-w-md mx-auto mt-16 px-4">
      <h1 className="text-2xl font-bold text-center mb-8">
        {t('注册', 'Register')}
      </h1>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        {error && (
          <div className="bg-red-50 text-red-600 text-sm px-4 py-3 rounded-lg">{error}</div>
        )}

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            {t('昵称', 'Nickname')}
          </label>
          <input
            type="text"
            {...register('nickname')}
            className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-rose-400"
            placeholder={t('请输入昵称', 'Enter your nickname')}
          />
          {errors.nickname && (
            <p className="text-red-500 text-xs mt-1">{errors.nickname.message}</p>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            {t('邮箱', 'Email')}
          </label>
          <input
            type="email"
            {...register('email')}
            className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-rose-400"
            placeholder={t('请输入邮箱', 'Enter your email')}
          />
          {errors.email && (
            <p className="text-red-500 text-xs mt-1">{errors.email.message}</p>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            {t('密码', 'Password')}
          </label>
          <input
            type="password"
            {...register('password')}
            className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-rose-400"
            placeholder={t('请输入密码（至少6位）', 'Enter your password (min 6 chars)')}
          />
          {errors.password && (
            <p className="text-red-500 text-xs mt-1">{errors.password.message}</p>
          )}
        </div>

        <button
          type="submit"
          disabled={loading}
          className="w-full bg-rose-500 text-white py-2 rounded-lg hover:bg-rose-600 disabled:opacity-50"
        >
          {loading ? t('注册中...', 'Registering...') : t('注册', 'Register')}
        </button>

        <p className="text-center text-sm text-gray-500">
          {t('已有账号？', 'Already have an account?')}{' '}
          <Link to="/login" className="text-rose-500 hover:underline">
            {t('登录', 'Login')}
          </Link>
        </p>
      </form>
    </div>
  )
}
