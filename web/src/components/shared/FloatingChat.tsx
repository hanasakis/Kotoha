import { useState, useRef, useEffect } from 'react'
import { useAuthStore } from '../../store/authStore'
import * as agentApi from '../../api/agent'
import type { ChatMessage } from '../../api/agent'

export default function FloatingChat() {
  const locale = useAuthStore((s) => s.locale)
  const isLoggedIn = useAuthStore((s) => !!s.accessToken)
  const [open, setOpen] = useState(false)
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [input, setInput] = useState('')
  const [sending, setSending] = useState(false)
  const bottomRef = useRef<HTMLDivElement>(null)

  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const handleOpen = () => {
    setOpen(true)
    if (messages.length === 0) {
      setMessages([
        {
          role: 'assistant',
          content: t(
            '你好！我是 Kotoha 的 AI 购物助手，可以帮你推荐零食、回答口味搭配问题。有什么我可以帮你的吗？',
            "Hi! I'm Kotoha's AI shopping assistant. I can help you find snacks and answer taste-matching questions. How can I help?"
          ),
        },
      ])
    }
  }

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

  if (!isLoggedIn) return null

  return (
    <>
      {/* FAB */}
      <button
        onClick={() => setOpen(!open)}
        className="fixed bottom-6 right-6 z-50 w-14 h-14 bg-rose-500 hover:bg-rose-600 text-white rounded-full shadow-lg flex items-center justify-center transition-all duration-200 hover:scale-105"
        title={t('AI 助手', 'AI Assistant')}
      >
        {open ? (
          <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        ) : (
          <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" />
          </svg>
        )}
      </button>

      {/* Chat panel */}
      {open && (
        <div className="fixed bottom-24 right-6 z-50 w-96 h-[520px] max-h-[calc(100vh-8rem)] bg-white rounded-2xl shadow-2xl border flex flex-col overflow-hidden">
          {/* Header */}
          <div className="px-4 py-3 bg-rose-500 text-white flex items-center justify-between">
            <div className="flex items-center gap-2">
              <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" />
              </svg>
              <span className="font-semibold text-sm">{t('AI 助手', 'AI Assistant')}</span>
            </div>
            <button onClick={() => setOpen(false)} className="text-white/80 hover:text-white">
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          {/* Messages */}
          <div className="flex-1 overflow-y-auto p-4 space-y-3 bg-gray-50">
            {messages.map((msg, i) => (
              <div key={i} className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}>
                <div
                  className={`max-w-[85%] px-3 py-2 rounded-xl text-sm whitespace-pre-wrap ${
                    msg.role === 'user' ? 'bg-rose-500 text-white' : 'bg-white border text-gray-700'
                  }`}
                >
                  {msg.content}
                </div>
              </div>
            ))}
            {sending && (
              <div className="flex justify-start">
                <div className="bg-white border px-3 py-2 rounded-xl text-sm text-gray-400">
                  {t('思考中...', 'Thinking...')}
                </div>
              </div>
            )}
            <div ref={bottomRef} />
          </div>

          {/* Input */}
          <div className="p-3 border-t bg-white">
            <div className="flex gap-2">
              <input
                type="text"
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder={t('输入你的问题...', 'Ask something...')}
                className="flex-1 px-3 py-2 text-sm border rounded-lg focus:outline-none focus:ring-2 focus:ring-rose-400"
                disabled={sending}
              />
              <button
                onClick={handleSend}
                disabled={!input.trim() || sending}
                className="px-4 py-2 bg-rose-500 text-white text-sm rounded-lg hover:bg-rose-600 disabled:opacity-50"
              >
                {t('发送', 'Send')}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}
