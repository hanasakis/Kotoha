import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useAuthStore } from '../../store/authStore'
import * as adminApi from '../../api/admin'
import type { EvalRun } from '../../types/admin'

function AccuracyBar({ value, max }: { value: number; max: number }) {
  const pct = Math.min((value / max) * 100, 100)
  const color = value >= 90 ? 'bg-green-500' : value >= 70 ? 'bg-yellow-500' : 'bg-red-500'
  return (
    <div className="w-full bg-gray-100 rounded-full h-2">
      <div className={`${color} h-2 rounded-full transition-all`} style={{ width: `${pct}%` }} />
    </div>
  )
}

function MetricsOverview() {
  const locale = useAuthStore((s) => s.locale)
  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)

  const { data, isLoading } = useQuery({
    queryKey: ['admin-metrics'],
    queryFn: adminApi.getMetrics,
  })

  if (isLoading) return <p className="text-sm text-gray-400">{t('加载中...', 'Loading...')}</p>
  if (!data?.metrics?.length) return <p className="text-sm text-gray-400">{t('暂无指标数据', 'No metrics data')}</p>

  return (
    <div className="bg-white border rounded-xl p-5">
      <h2 className="font-semibold text-gray-900 mb-4">{t('业务指标', 'Business Metrics')}</h2>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {data.metrics.map((m) => (
          <div key={m.category} className="border rounded-lg p-4">
            <h3 className="font-medium text-sm text-gray-700 mb-3">{m.category}</h3>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-gray-500">{t('搜索次数', 'Searches')}</span>
                <span className="font-medium">{m.total_searches}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">{t('平均相关度', 'Avg Relevance')}</span>
                <span className="font-medium">{m.avg_relevance.toFixed(2)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">{t('加购数', 'Cart Adds')}</span>
                <span className="font-medium">{m.cart_adds}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">{t('下单数', 'Orders')}</span>
                <span className="font-medium">{m.orders_created}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-gray-500">{t('转化率', 'Conversion')}</span>
                <span className="font-medium">
                  {m.total_searches > 0 ? ((m.orders_created / m.total_searches) * 100).toFixed(1) : 0}%
                </span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

const categoryOptions = [
  { value: '', label: 'All', labelZh: '全部' },
  { value: 'search_relevance', label: 'Search Relevance', labelZh: '搜索相关度' },
  { value: 'agent_accuracy', label: 'Agent Accuracy', labelZh: 'Agent 准确度' },
  { value: 'recommendation', label: 'Recommendation', labelZh: '推荐质量' },
]

export default function AdminEvaluationsPage() {
  const locale = useAuthStore((s) => s.locale)
  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)
  const [category, setCategory] = useState('')

  const { data: evalRuns, isLoading } = useQuery({
    queryKey: ['admin-eval-runs', category],
    queryFn: () => adminApi.listEvalRuns(category || undefined, 50),
  })

  const maxAccuracy = Math.max(...(evalRuns || []).map((r) => r.accuracy), 100)

  return (
    <div className="space-y-6">
      <h1 className="text-xl font-bold text-gray-900">{t('评测分析', 'Evaluations')}</h1>

      {/* Metrics Overview */}
      <MetricsOverview />

      {/* Eval Runs */}
      <div className="bg-white border rounded-xl p-5">
        <div className="flex justify-between items-center mb-4">
          <h2 className="font-semibold text-gray-900">{t('评测运行记录', 'Evaluation Runs')}</h2>
          <select
            value={category}
            onChange={(e) => setCategory(e.target.value)}
            className="border rounded-lg px-3 py-1.5 text-sm"
          >
            {categoryOptions.map((c) => (
              <option key={c.value} value={c.value}>{t(c.labelZh, c.label)}</option>
            ))}
          </select>
        </div>

        {isLoading ? (
          <p className="text-sm text-gray-400 py-8 text-center">{t('加载中...', 'Loading...')}</p>
        ) : !evalRuns?.length ? (
          <p className="text-sm text-gray-400 py-8 text-center">{t('暂无评测记录', 'No evaluation runs yet. Run evals from the backend to see results.')}</p>
        ) : (
          <div className="space-y-3">
            {evalRuns.map((run: EvalRun) => (
              <div key={run.id} className="border rounded-lg p-4">
                <div className="flex justify-between items-start mb-2">
                  <div>
                    <span className="text-sm font-medium text-gray-800">{run.category}</span>
                    <span className="text-xs text-gray-400 ml-2">
                      {new Date(run.created_at).toLocaleString()}
                    </span>
                  </div>
                  <span className={`text-sm font-bold ${
                    run.accuracy >= 90 ? 'text-green-600' : run.accuracy >= 70 ? 'text-yellow-600' : 'text-red-500'
                  }`}>
                    {run.accuracy.toFixed(1)}%
                  </span>
                </div>
                <AccuracyBar value={run.accuracy} max={maxAccuracy} />
                <div className="flex gap-4 mt-2 text-xs text-gray-500">
                  <span>{t('总计', 'Total')}: {run.total_tests}</span>
                  <span className="text-green-600">{t('通过', 'Passed')}: {run.passed}</span>
                  <span className="text-red-500">{t('失败', 'Failed')}: {run.failed}</span>
                  <span>{t('通过率', 'Pass rate')}: {run.total_tests > 0 ? ((run.passed / run.total_tests) * 100).toFixed(1) : 0}%</span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
