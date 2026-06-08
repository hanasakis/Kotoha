import { useState, useRef, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../store/authStore'
import { useCartStore } from '../../store/cartStore'

function DropdownMenu({ label, locale, onLogout }: { label: string; locale: string; onLogout: () => void }) {
  const [open, setOpen] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)
  const closeTimer = useRef<ReturnType<typeof setTimeout>>()

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const handleMouseEnter = () => {
    clearTimeout(closeTimer.current)
    setOpen(true)
  }

  const handleMouseLeave = () => {
    closeTimer.current = setTimeout(() => setOpen(false), 200)
  }

  const handleLinkClick = () => setOpen(false)

  return (
    <div
      ref={containerRef}
      className="relative"
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
    >
      <button className="text-gray-600 hover:text-gray-900">
        {label}
      </button>
      {open && (
        <div className="absolute right-0 pt-2 z-50">
          <div className="w-40 bg-white border rounded-lg shadow-lg">
            <Link to="/profile" onClick={handleLinkClick} className="block px-4 py-2 text-sm hover:bg-gray-50">
              {locale === 'zh' ? '个人中心' : 'Profile'}
            </Link>
            <Link to="/profile/addresses" onClick={handleLinkClick} className="block px-4 py-2 text-sm hover:bg-gray-50">
              {locale === 'zh' ? '收货地址' : 'Addresses'}
            </Link>
            <button onClick={() => { handleLinkClick(); onLogout() }} className="block w-full text-left px-4 py-2 text-sm text-red-500 hover:bg-gray-50">
              {locale === 'zh' ? '退出登录' : 'Logout'}
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

export default function Header() {
  const [query, setQuery] = useState('')
  const navigate = useNavigate()
  const { isAuthenticated, user, locale } = useAuthStore()
  const setLocale = useAuthStore((s) => s.setLocale)
  const logout = useAuthStore((s) => s.logout)
  const items = useCartStore((s) => s.items)
  const itemCount = items.reduce((sum, i) => sum + i.quantity, 0)

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    if (query.trim()) navigate(`/search?q=${encodeURIComponent(query.trim())}`)
  }

  return (
    <header className="bg-white shadow-sm border-b sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 h-16 flex items-center gap-4">
        <Link to="/" className="text-xl font-bold text-rose-600 shrink-0">
          Kotoha
        </Link>

        <form onSubmit={handleSearch} className="flex-1 max-w-md">
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={locale === 'zh' ? '搜索零食...' : 'Search snacks...'}
            className="w-full px-3 py-1.5 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-rose-400"
          />
        </form>

        <nav className="flex items-center gap-3 text-sm">
          <Link to="/products" className="text-gray-600 hover:text-gray-900">
            {locale === 'zh' ? '全部商品' : 'Products'}
          </Link>

          {isAuthenticated ? (
            <>
              <Link to="/cart" className="relative text-gray-600 hover:text-gray-900">
                {locale === 'zh' ? '购物车' : 'Cart'}
                {itemCount > 0 && (
                  <span className="absolute -top-2 -right-4 bg-rose-500 text-white text-xs rounded-full w-5 h-5 flex items-center justify-center">
                    {itemCount}
                  </span>
                )}
              </Link>
              <Link to="/orders" className="text-gray-600 hover:text-gray-900">
                {locale === 'zh' ? '订单' : 'Orders'}
              </Link>
              <Link to="/agent" className="text-gray-600 hover:text-gray-900">
                AI
              </Link>
              {user?.role === 'admin' && (
                <Link to="/admin" className="text-rose-500 hover:text-rose-700 font-medium">
                  {locale === 'zh' ? '管理' : 'Admin'}
                </Link>
              )}
              <DropdownMenu
                label={user?.nickname || user?.email || (locale === 'zh' ? '我的' : 'Me')}
                locale={locale}
                onLogout={logout}
              />
            </>
          ) : (
            <>
              <Link to="/login" className="text-gray-600 hover:text-gray-900">
                {locale === 'zh' ? '登录' : 'Login'}
              </Link>
              <Link to="/register" className="bg-rose-500 text-white px-3 py-1 rounded-lg hover:bg-rose-600">
                {locale === 'zh' ? '注册' : 'Register'}
              </Link>
            </>
          )}

          <button
            onClick={() => setLocale(locale === 'zh' ? 'en' : 'zh')}
            className="text-xs text-gray-400 hover:text-gray-600 ml-2"
          >
            {locale === 'zh' ? 'EN' : '中'}
          </button>
        </nav>
      </div>
    </header>
  )
}
