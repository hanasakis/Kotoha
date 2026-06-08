import { Navigate, Outlet } from 'react-router-dom'
import { useAuthStore } from '../../store/authStore'

export default function AdminGuard() {
  const { isAuthenticated, user } = useAuthStore()

  if (!isAuthenticated || !user) {
    return <Navigate to="/login?redirect=/admin" replace />
  }

  if (user.role !== 'admin') {
    return (
      <div className="max-w-md mx-auto mt-32 px-4 text-center">
        <h1 className="text-6xl font-bold text-gray-200 mb-4">403</h1>
        <p className="text-gray-500 mb-6">You do not have permission to access this page.</p>
        <a href="/" className="text-rose-500 hover:underline">Back to Home</a>
      </div>
    )
  }

  return <Outlet />
}
