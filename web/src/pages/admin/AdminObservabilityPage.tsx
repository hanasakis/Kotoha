import axios from 'axios'
import { useQuery } from '@tanstack/react-query'
import { useAuthStore } from '../../store/authStore'
import apiClient from '../../api/client'

interface HealthStatus {
  status: string
  deps: Record<string, string>
}

function StatusDot({ status }: { status: string }) {
  const color = status === 'ok' ? 'bg-green-500' : status === 'degraded' ? 'bg-yellow-500' : 'bg-red-500'
  return <span className={`inline-block w-2.5 h-2.5 rounded-full ${color} mr-2`} />
}

export default function AdminObservabilityPage() {
  const locale = useAuthStore((s) => s.locale)
  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)

  const { data: health, isLoading: healthLoading } = useQuery<HealthStatus>({
    queryKey: ['health'],
    queryFn: async () => {
      const { data } = await axios.get('/health')
      return data.data
    },
    refetchInterval: 30000,
  })

  return (
    <div className="space-y-6">
      <h1 className="text-xl font-bold text-gray-900">{t('可观测性', 'Observability')}</h1>

      {/* System Health */}
      <div className="bg-white border rounded-xl p-5">
        <h2 className="font-semibold text-gray-900 mb-4">{t('系统健康状态', 'System Health')}</h2>
        {healthLoading ? (
          <p className="text-sm text-gray-400">{t('检查中...', 'Checking...')}</p>
        ) : health ? (
          <div className="space-y-3">
            <div className="flex items-center gap-3 mb-4">
              <span className="text-sm text-gray-500">{t('整体状态', 'Overall')}:</span>
              <span className={`px-3 py-1 rounded-full text-xs font-medium ${
                health.status === 'ok' ? 'bg-green-100 text-green-700' :
                health.status === 'degraded' ? 'bg-yellow-100 text-yellow-700' :
                'bg-red-100 text-red-700'
              }`}>
                {health.status}
              </span>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {Object.entries(health.deps || {}).map(([name, status]) => (
                <div key={name} className="flex items-center justify-between border rounded-lg px-4 py-3">
                  <div className="flex items-center">
                    <StatusDot status={status} />
                    <span className="text-sm font-medium text-gray-700">{name}</span>
                  </div>
                  <span className={`text-xs px-2 py-0.5 rounded-full ${
                    status === 'ok' ? 'bg-green-100 text-green-700' :
                    status.includes('unhealthy') ? 'bg-red-100 text-red-700' :
                    'bg-yellow-100 text-yellow-700'
                  }`}>
                    {status}
                  </span>
                </div>
              ))}
            </div>
          </div>
        ) : (
          <p className="text-sm text-red-500">{t('无法获取健康状态', 'Unable to fetch health status')}</p>
        )}
      </div>

      {/* Langfuse */}
      <div className="bg-white border rounded-xl p-5">
        <h2 className="font-semibold text-gray-900 mb-3">Langfuse</h2>
        <p className="text-sm text-gray-500 mb-4">
          {t(
            'Langfuse 用于 LLM 可观测性：追踪 Agent 对话、搜索查询、购物车操作和订单创建。',
            'Langfuse provides LLM observability: tracing agent conversations, search queries, cart actions, and order creation.'
          )}
        </p>
        <div className="space-y-2 text-sm">
          <div className="flex justify-between items-center py-2 border-t">
            <span className="text-gray-600">{t('Agent 对话追踪', 'Agent Chat Tracing')}</span>
            <span className="text-xs px-2 py-0.5 rounded-full bg-blue-100 text-blue-700">{t('已集成', 'Integrated')}</span>
          </div>
          <div className="flex justify-between items-center py-2 border-t">
            <span className="text-gray-600">{t('搜索追踪', 'Search Tracing')}</span>
            <span className="text-xs px-2 py-0.5 rounded-full bg-blue-100 text-blue-700">{t('已集成', 'Integrated')}</span>
          </div>
          <div className="flex justify-between items-center py-2 border-t">
            <span className="text-gray-600">{t('购物车追踪', 'Cart Action Tracing')}</span>
            <span className="text-xs px-2 py-0.5 rounded-full bg-blue-100 text-blue-700">{t('已集成', 'Integrated')}</span>
          </div>
          <div className="flex justify-between items-center py-2 border-t">
            <span className="text-gray-600">{t('订单追踪', 'Order Tracing')}</span>
            <span className="text-xs px-2 py-0.5 rounded-full bg-blue-100 text-blue-700">{t('已集成', 'Integrated')}</span>
          </div>
        </div>
        <p className="text-xs text-gray-400 mt-4">
          {t('查看 Langfuse 控制台以获取详细的 trace 数据。', 'Check the Langfuse console for detailed trace data.')}
        </p>
      </div>

      {/* Stripe */}
      <div className="bg-white border rounded-xl p-5">
        <h2 className="font-semibold text-gray-900 mb-3">Stripe</h2>
        <p className="text-sm text-gray-500 mb-4">
          {t(
            'Stripe 处理支付流程：Webhook 接收支付事件，自动更新订单状态。',
            'Stripe handles payment processing: webhooks receive payment events and auto-update order status.'
          )}
        </p>
        <div className="space-y-2 text-sm">
          <div className="flex justify-between items-center py-2 border-t">
            <span className="text-gray-600">{t('Checkout Session', 'Checkout Session')}</span>
            <span className="text-xs px-2 py-0.5 rounded-full bg-blue-100 text-blue-700">{t('已集成', 'Integrated')}</span>
          </div>
          <div className="flex justify-between items-center py-2 border-t">
            <span className="text-gray-600">{t('Webhook 处理', 'Webhook Handler')}</span>
            <span className="text-xs px-2 py-0.5 rounded-full bg-blue-100 text-blue-700">POST /api/v1/webhook</span>
          </div>
          <div className="flex justify-between items-center py-2 border-t">
            <span className="text-gray-600">{t('退款流程', 'Refund Flow')}</span>
            <span className="text-xs px-2 py-0.5 rounded-full bg-blue-100 text-blue-700">{t('已集成', 'Integrated')}</span>
          </div>
        </div>
        <p className="text-xs text-gray-400 mt-4">
          {t(
            'Webhook 端点需在 Stripe Dashboard 中配置。支付事件：checkout.session.completed, payment_intent.payment_failed。',
            'Configure the webhook endpoint in Stripe Dashboard. Payment events: checkout.session.completed, payment_intent.payment_failed.'
          )}
        </p>
      </div>
    </div>
  )
}
