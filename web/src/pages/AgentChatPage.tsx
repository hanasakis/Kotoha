import { useState, useRef, useEffect } from 'react'
import { useAuthStore } from '../store/authStore'
import * as agentApi from '../api/agent'
import type { ChatMessage } from '../api/agent'

export default function AgentChatPage() {
  const locale = useAuthStore((s) => s.locale)
  const [messages, setMessages] = useState<ChatMessage[]>([
    { role: 'assistant', content: locale === 'zh' ? '你好！我是 Kotoha 的 AI 购物助手，可以帮你推荐零食、回答口味搭配问题。有什么我可以帮你的吗？' : 'Hi! I\'m Kotoha\'s AI shopping assistant. I can help you find snacks and answer taste-matching questions. How can I help?' },
  ])
  const [input, setInput] = useState('')
  const [sending, setSending] = useState(false)
  const bottomRef = useRef<HTMLDivElement>(null)

  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const handleSend = async () => {
    if (!input.trim() || sending) return
    const userMsg: ChatMessage = { role: 'user', content: input.trim() }
    setMessages((prev) => [...prev, userMsg])
    setInput('')
    setSending(true)

    try {
      const response = await agentApi.sendChat({
        message: userMsg.content,
        history: messages.slice(-10),
      })
      setMessages((prev) => [...prev, { role: 'assistant', content: response.reply }])
    } catch {
      setMessages((prev) => [
        ...prev,
        { role: 'assistant', content: t('抱歉，出了点问题，请稍后重试。', 'Sorry, something went wrong. Please try again later.') },
      ])
    } finally {
      setSending(false)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  return (
    <div className="max-w-2xl mx-auto px-4 py-6 h-[calc(100vh-8rem)] flex flex-col">
      <h1 className="text-2xl font-bold mb-4">{t('AI 助手', 'AI Assistant')}</h1>

      <div className="flex-1 overflow-y-auto space-y-3 mb-4 bg-gray-50 rounded-xl p-4">
        {messages.map((msg, i) => (
          <div
            key={i}
            className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}
          >
            <div
              className={`max-w-[80%] px-4 py-2 rounded-xl text-sm whitespace-pre-wrap ${
                msg.role === 'user'
                  ? 'bg-rose-500 text-white'
                  : 'bg-white border text-gray-700'
              }`}
            >
              {msg.content}
            </div>
          </div>
        ))}
        {sending && (
          <div className="flex justify-start">
            <div className="bg-white border px-4 py-2 rounded-xl text-sm text-gray-400">
              {t('思考中...', 'Thinking...')}
            </div>
          </div>
        )}
        <div ref={bottomRef} />
      </div>

      <div className="flex gap-2">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={t('输入你的问题...', 'Ask something...')}
          className="flex-1 px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-rose-400"
          disabled={sending}
        />
        <button
          onClick={handleSend}
          disabled={!input.trim() || sending}
          className="px-6 py-2 bg-rose-500 text-white rounded-lg hover:bg-rose-600 disabled:opacity-50"
        >
          {t('发送', 'Send')}
        </button>
      </div>
    </div>
  )
}
