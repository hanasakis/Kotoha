import { Outlet } from 'react-router-dom'
import Header from './Header'
import Footer from './Footer'
import FloatingChat from '../shared/FloatingChat'

export default function AppLayout() {
  return (
    <div className="min-h-screen flex flex-col bg-gray-50">
      <Header />
      <main className="flex-1 w-full max-w-7xl mx-auto px-4 py-6">
        <Outlet />
      </main>
      <Footer />
      <FloatingChat />
    </div>
  )
}
