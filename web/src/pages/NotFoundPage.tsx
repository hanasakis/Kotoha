import { Link } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'

export default function NotFoundPage() {
  const locale = useAuthStore((s) => s.locale)
  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)

  return (
    <div className="max-w-md mx-auto mt-16 px-4 text-center">
      <h1 className="text-6xl font-bold text-gray-200 mb-4">404</h1>
      <p className="text-gray-500 mb-6">
        {t('页面不存在', 'Page not found')}
      </p>
      <Link to="/" className="text-rose-500 hover:underline">
        {t('返回首页', 'Back to Home')}
      </Link>
    </div>
  )
}
