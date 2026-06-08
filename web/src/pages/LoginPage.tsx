import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'

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
  password: z.string().min(6, '密码至少6位'),
})

type FormData = z.infer<typeof schema>

export default function LoginPage() {
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const login = useAuthStore((s) => s.login)
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
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
      await login(data.email, data.password)
      const redirect = searchParams.get('redirect') || '/'
      navigate(redirect, { replace: true })
    } catch (err: unknown) {
      const e = err as { code?: string; message?: string }
      const errorMsg = locale === 'zh' ? '邮箱或密码错误' : 'Invalid email or password'
      setError(e.code === 'auth.login_failed' ? errorMsg : (e.message || errorMsg))
    } finally {
      setLoading(false)
    }
  }

  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)

  return (
    <div className="max-w-md mx-auto mt-16 px-4">
      <h1 className="text-2xl font-bold text-center mb-8">
        {t('登录', 'Login')}
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
            placeholder={t('请输入密码', 'Enter your password')}
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
          {loading ? t('登录中...', 'Logging in...') : t('登录', 'Login')}
        </button>

        <div className="flex justify-between text-sm">
          <Link to="/register" className="text-rose-500 hover:underline">
            {t('没有账号？注册', "Don't have an account? Register")}
          </Link>
          <Link to="/forgot-password" className="text-gray-500 hover:underline">
            {t('忘记密码？', 'Forgot password?')}
          </Link>
        </div>
      </form>
    </div>
  )
}
