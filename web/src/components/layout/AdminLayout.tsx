import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../store/authStore'

const navItems = [
  { to: '/admin', label: 'Dashboard', labelZh: '仪表盘', end: true },
  { to: '/admin/products', label: 'Products', labelZh: '商品管理' },
  { to: '/admin/orders', label: 'Orders', labelZh: '订单管理' },
  { to: '/admin/evaluations', label: 'Evaluations', labelZh: '评测分析' },
  { to: '/admin/observability', label: 'Observability', labelZh: '可观测性' },
]

export default function AdminLayout() {
  const { locale, logout } = useAuthStore()
  const navigate = useNavigate()
  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)

  const handleLogout = async () => {
    await logout()
    navigate('/')
  }

  return (
    <div className="min-h-screen flex bg-gray-50">
      {/* Sidebar */}
      <aside className="w-56 bg-white border-r shrink-0 flex flex-col">
        <div className="px-5 py-4 border-b">
          <NavLink to="/" className="text-lg font-bold text-rose-600">
            Kotoha
          </NavLink>
          <p className="text-xs text-gray-400 mt-0.5">{t('商家后台', 'Admin Panel')}</p>
        </div>
        <nav className="flex-1 px-3 py-4 space-y-1">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              className={({ isActive }) =>
                `block px-3 py-2 rounded-lg text-sm transition-colors ${
                  isActive
                    ? 'bg-rose-50 text-rose-600 font-medium'
                    : 'text-gray-600 hover:bg-gray-100'
                }`
              }
            >
              {t(item.labelZh, item.label)}
            </NavLink>
          ))}
        </nav>
        <div className="px-3 py-4 border-t">
          <NavLink
            to="/"
            className="block px-3 py-2 rounded-lg text-sm text-gray-500 hover:bg-gray-100 mb-1"
          >
            {t('← 返回商城', '← Back to Store')}
          </NavLink>
          <button
            onClick={handleLogout}
            className="block w-full text-left px-3 py-2 rounded-lg text-sm text-red-500 hover:bg-red-50"
          >
            {t('退出登录', 'Logout')}
          </button>
        </div>
      </aside>

      {/* Main content */}
      <div className="flex-1 min-w-0">
        <main className="p-6">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
