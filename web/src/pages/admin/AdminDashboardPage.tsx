import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { useAuthStore } from '../../store/authStore'
import * as adminApi from '../../api/admin'
import * as catalogApi from '../../api/catalog'
import type { EvalRun } from '../../types/admin'

function StatCard({ title, value, href, color }: { title: string; value: string | number; href?: string; color?: string }) {
  const body = (
    <div className={`bg-white border rounded-xl p-5 ${href ? 'hover:shadow-md transition-shadow cursor-pointer' : ''}`}>
      <p className="text-sm text-gray-500 mb-1">{title}</p>
      <p className={`text-2xl font-bold ${color || 'text-gray-900'}`}>{value}</p>
    </div>
  )
  return href ? <Link to={href}>{body}</Link> : body
}

function EvalRunRow({ run }: { run: EvalRun }) {
  const accuracyColor = run.accuracy >= 90 ? 'text-green-600' : run.accuracy >= 70 ? 'text-yellow-600' : 'text-red-500'
  return (
    <tr className="border-t hover:bg-gray-50">
      <td className="py-2.5 px-4 text-sm text-gray-600">{run.category}</td>
      <td className="py-2.5 px-4 text-sm text-center">{run.total_tests}</td>
      <td className="py-2.5 px-4 text-sm text-center text-green-600">{run.passed}</td>
      <td className="py-2.5 px-4 text-sm text-center text-red-500">{run.failed}</td>
      <td className={`py-2.5 px-4 text-sm text-center font-medium ${accuracyColor}`}>{run.accuracy.toFixed(1)}%</td>
      <td className="py-2.5 px-4 text-sm text-gray-400">{new Date(run.created_at).toLocaleDateString()}</td>
    </tr>
  )
}

export default function AdminDashboardPage() {
  const locale = useAuthStore((s) => s.locale)
  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)

  const { data: productsData } = useQuery({
    queryKey: ['admin-products-count'],
    queryFn: () => catalogApi.getProducts(1, 1),
  })

  const { data: ordersData } = useQuery({
    queryKey: ['admin-orders-count'],
    queryFn: () => adminApi.listAllOrders(1, 1),
  })

  const { data: metrics, isLoading: metricsLoading } = useQuery({
    queryKey: ['admin-metrics'],
    queryFn: adminApi.getMetrics,
  })

  const { data: evalRuns, isLoading: runsLoading } = useQuery({
    queryKey: ['admin-eval-runs', 5],
    queryFn: () => adminApi.listEvalRuns(undefined, 5),
  })

  return (
    <div>
      <h1 className="text-xl font-bold text-gray-900 mb-6">{t('仪表盘', 'Dashboard')}</h1>

      {/* Stats Grid */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <StatCard title={t('商品总数', 'Total Products')} value={productsData?.total || '-'} href="/admin/products" color="text-blue-600" />
        <StatCard title={t('订单总数', 'Total Orders')} value={ordersData?.total || '-'} href="/admin/orders" color="text-green-600" />
        <StatCard title={t('评测指标', 'Eval Metrics')} value={metricsLoading ? '...' : (metrics?.metrics?.length || 0)} href="/admin/evaluations" color="text-purple-600" />
        <StatCard title={t('Langfuse', 'Langfuse')} value="—" href="/admin/observability" color="text-orange-600" />
      </div>

      {/* Quick Links */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Quick Actions */}
        <div className="bg-white border rounded-xl p-5">
          <h2 className="font-semibold text-gray-900 mb-4">{t('快捷操作', 'Quick Actions')}</h2>
          <div className="space-y-2">
            <Link to="/admin/products" className="block px-4 py-2.5 bg-gray-50 rounded-lg text-sm text-gray-700 hover:bg-rose-50 hover:text-rose-600 transition-colors">
              {t('管理商品 →', 'Manage Products →')}
            </Link>
            <Link to="/admin/orders" className="block px-4 py-2.5 bg-gray-50 rounded-lg text-sm text-gray-700 hover:bg-rose-50 hover:text-rose-600 transition-colors">
              {t('查看订单 →', 'View Orders →')}
            </Link>
            <Link to="/admin/evaluations" className="block px-4 py-2.5 bg-gray-50 rounded-lg text-sm text-gray-700 hover:bg-rose-50 hover:text-rose-600 transition-colors">
              {t('评测分析 →', 'Evaluation Analysis →')}
            </Link>
            <Link to="/admin/observability" className="block px-4 py-2.5 bg-gray-50 rounded-lg text-sm text-gray-700 hover:bg-rose-50 hover:text-rose-600 transition-colors">
              {t('可观测性 →', 'Observability →')}
            </Link>
          </div>
        </div>

        {/* Recent Eval Runs */}
        <div className="bg-white border rounded-xl p-5">
          <div className="flex justify-between items-center mb-4">
            <h2 className="font-semibold text-gray-900">{t('最近评测', 'Recent Evaluations')}</h2>
            <Link to="/admin/evaluations" className="text-xs text-rose-500 hover:underline">
              {t('查看全部', 'View All')} →
            </Link>
          </div>
          {runsLoading ? (
            <p className="text-sm text-gray-400">{t('加载中...', 'Loading...')}</p>
          ) : !evalRuns?.length ? (
            <p className="text-sm text-gray-400">{t('暂无评测数据', 'No evaluation data yet')}</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-left text-gray-500">
                    <th className="py-2 px-4 font-medium">{t('类别', 'Category')}</th>
                    <th className="py-2 px-4 font-medium text-center">{t('通过', 'Pass')}</th>
                    <th className="py-2 px-4 font-medium text-center">{t('失败', 'Fail')}</th>
                    <th className="py-2 px-4 font-medium text-center">{t('准确率', 'Accuracy')}</th>
                  </tr>
                </thead>
                <tbody>
                  {evalRuns.map((run) => (
                    <EvalRunRow key={run.id} run={run} />
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      {/* Metrics Overview */}
      {metrics?.metrics && metrics.metrics.length > 0 && (
        <div className="mt-6 bg-white border rounded-xl p-5">
          <h2 className="font-semibold text-gray-900 mb-4">{t('指标概览', 'Metrics Overview')}</h2>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-gray-500 border-b">
                  <th className="py-2 px-4 font-medium">{t('类别', 'Category')}</th>
                  <th className="py-2 px-4 font-medium text-center">{t('搜索次数', 'Searches')}</th>
                  <th className="py-2 px-4 font-medium text-center">{t('平均相关度', 'Avg Relevance')}</th>
                  <th className="py-2 px-4 font-medium text-center">{t('加购数', 'Cart Adds')}</th>
                  <th className="py-2 px-4 font-medium text-center">{t('下单数', 'Orders')}</th>
                </tr>
              </thead>
              <tbody>
                {metrics.metrics.map((m) => (
                  <tr key={m.category} className="border-t hover:bg-gray-50">
                    <td className="py-2.5 px-4 text-gray-700">{m.category}</td>
                    <td className="py-2.5 px-4 text-center text-gray-600">{m.total_searches}</td>
                    <td className="py-2.5 px-4 text-center text-gray-600">{m.avg_relevance.toFixed(2)}</td>
                    <td className="py-2.5 px-4 text-center text-gray-600">{m.cart_adds}</td>
                    <td className="py-2.5 px-4 text-center text-gray-600">{m.orders_created}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}
