import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useAuthStore } from '../store/authStore'
import * as userApi from '../api/user'
import * as authApi from '../api/auth'

export default function ProfilePage() {
  const locale = useAuthStore((s) => s.locale)
  const user = useAuthStore((s) => s.user)
  const logout = useAuthStore((s) => s.logout)
  const queryClient = useQueryClient()

  const [phone, setPhone] = useState('')
  const [saving, setSaving] = useState(false)
  const [saveMsg, setSaveMsg] = useState('')
  const [deletePwd, setDeletePwd] = useState('')
  const [deleting, setDeleting] = useState(false)
  const [deleteError, setDeleteError] = useState('')

  const [prefDietary, setPrefDietary] = useState('')
  const [prefTaste, setPrefTaste] = useState('')
  const [prefScenes, setPrefScenes] = useState('')
  const [prefAllergens, setPrefAllergens] = useState('')
  const [prefSaving, setPrefSaving] = useState(false)
  const [prefMsg, setPrefMsg] = useState('')

  const t = (zh: string, en: string) => (locale === 'zh' ? zh : en)

  const { data: profile } = useQuery({
    queryKey: ['profile'],
    queryFn: () => userApi.getProfile().then((p) => { setPhone(p.phone || ''); return p }),
  })

  const { data: preference } = useQuery({
    queryKey: ['preference'],
    queryFn: () => userApi.getPreference().then((p) => {
      if (p) {
        setPrefDietary(p.dietary_limits || '')
        setPrefTaste(p.taste_prefs || '')
        setPrefScenes(p.scene_prefs || '')
        setPrefAllergens(p.allergens || '')
      }
      return p
    }),
  })

  const handleSavePrefs = async () => {
    setPrefSaving(true)
    setPrefMsg('')
    try {
      await userApi.updatePreference({
        dietary_limits: prefDietary || undefined,
        taste_prefs: prefTaste || undefined,
        scene_prefs: prefScenes || undefined,
        allergens: prefAllergens || undefined,
      })
      queryClient.invalidateQueries({ queryKey: ['preference'] })
      setPrefMsg(t('保存成功', 'Saved'))
    } catch {
      setPrefMsg(t('保存失败', 'Save failed'))
    } finally {
      setPrefSaving(false)
    }
  }

  const handleSaveProfile = async () => {
    setSaving(true)
    setSaveMsg('')
    try {
      await userApi.updateProfile({ phone: phone || undefined })
      setSaveMsg(t('保存成功', 'Saved'))
    } catch {
      setSaveMsg(t('保存失败', 'Save failed'))
    } finally {
      setSaving(false)
    }
  }

  const handleDeleteAccount = async () => {
    if (!deletePwd) return
    if (!confirm(t('确定要注销账号吗？此操作不可恢复。', 'Are you sure? This cannot be undone.'))) return
    setDeleting(true)
    setDeleteError('')
    try {
      await authApi.deleteAccount({ password: deletePwd })
      logout()
    } catch (err: unknown) {
      const e = err as { code?: string; message?: string }
      setDeleteError(e.message || t('注销失败', 'Deletion failed'))
    } finally {
      setDeleting(false)
    }
  }

  if (!user) return null

  return (
    <div className="max-w-2xl mx-auto px-4 py-6 space-y-8">
      <h1 className="text-2xl font-bold">{t('个人中心', 'Profile')}</h1>

      {/* Basic info */}
      <section className="bg-white border rounded-xl p-6 space-y-4">
        <h2 className="font-bold">{t('基本信息', 'Basic Info')}</h2>
        <div>
          <label className="text-sm text-gray-500">{t('邮箱', 'Email')}</label>
          <p className="text-sm">{user.email}</p>
        </div>
        <div>
          <label className="text-sm text-gray-500">{t('昵称', 'Nickname')}</label>
          <p className="text-sm">{user.nickname || '-'}</p>
        </div>
        <div>
          <label className="text-sm text-gray-500">{t('角色', 'Role')}</label>
          <p className="text-sm">{user.role}</p>
        </div>
        <div>
          <label className="text-sm text-gray-500">{t('手机号', 'Phone')}</label>
          <input
            type="text"
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            className="w-full mt-1 px-3 py-1.5 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-rose-400"
          />
        </div>
        <button
          onClick={handleSaveProfile}
          disabled={saving}
          className="px-4 py-1.5 bg-rose-500 text-white text-sm rounded-lg hover:bg-rose-600 disabled:opacity-50"
        >
          {saving ? t('保存中...', 'Saving...') : t('保存', 'Save')}
        </button>
        {saveMsg && <span className="ml-3 text-sm text-green-600">{saveMsg}</span>}
      </section>

      {/* Preferences */}
      <section className="bg-white border rounded-xl p-6 space-y-3">
        <h2 className="font-bold">{t('偏好设置', 'Preferences')}</h2>
        <div>
          <label className="text-sm text-gray-500">{t('饮食限制', 'Dietary')}</label>
          <input
            type="text"
            value={prefDietary}
            onChange={(e) => setPrefDietary(e.target.value)}
            placeholder={t('如：素食、清真', 'e.g. vegetarian, halal')}
            className="w-full mt-1 px-3 py-1.5 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-rose-400"
          />
        </div>
        <div>
          <label className="text-sm text-gray-500">{t('口味偏好', 'Taste')}</label>
          <input
            type="text"
            value={prefTaste}
            onChange={(e) => setPrefTaste(e.target.value)}
            placeholder={t('如：辣、甜、咸', 'e.g. spicy, sweet, salty')}
            className="w-full mt-1 px-3 py-1.5 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-rose-400"
          />
        </div>
        <div>
          <label className="text-sm text-gray-500">{t('场景偏好', 'Scenes')}</label>
          <input
            type="text"
            value={prefScenes}
            onChange={(e) => setPrefScenes(e.target.value)}
            placeholder={t('如：追剧、办公、聚会', 'e.g. movie night, office, party')}
            className="w-full mt-1 px-3 py-1.5 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-rose-400"
          />
        </div>
        <div>
          <label className="text-sm text-gray-500">{t('过敏原', 'Allergens')}</label>
          <input
            type="text"
            value={prefAllergens}
            onChange={(e) => setPrefAllergens(e.target.value)}
            placeholder={t('如：花生、牛奶、海鲜', 'e.g. peanuts, milk, seafood')}
            className="w-full mt-1 px-3 py-1.5 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-rose-400"
          />
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={handleSavePrefs}
            disabled={prefSaving}
            className="px-4 py-1.5 bg-rose-500 text-white text-sm rounded-lg hover:bg-rose-600 disabled:opacity-50"
          >
            {prefSaving ? t('保存中...', 'Saving...') : t('保存偏好', 'Save Preferences')}
          </button>
          {prefMsg && <span className="text-sm text-green-600">{prefMsg}</span>}
        </div>
      </section>

      {/* Danger zone */}
      <section className="bg-white border border-red-200 rounded-xl p-6 space-y-4">
        <h2 className="font-bold text-red-500">{t('注销账号', 'Delete Account')}</h2>
        <p className="text-sm text-gray-500">
          {t('注销后账号将被永久删除，数据无法恢复。', 'Your account and all data will be permanently deleted.')}
        </p>
        {deleteError && (
          <div className="bg-red-50 text-red-600 text-sm px-4 py-3 rounded-lg">{deleteError}</div>
        )}
        <div className="flex gap-2">
          <input
            type="password"
            value={deletePwd}
            onChange={(e) => setDeletePwd(e.target.value)}
            placeholder={t('请输入密码确认', 'Enter password to confirm')}
            className="flex-1 px-3 py-1.5 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-red-400"
          />
          <button
            onClick={handleDeleteAccount}
            disabled={deleting || !deletePwd}
            className="px-4 py-1.5 bg-red-500 text-white text-sm rounded-lg hover:bg-red-600 disabled:opacity-50"
          >
            {deleting ? t('注销中...', 'Deleting...') : t('确认注销', 'Delete')}
          </button>
        </div>
      </section>
    </div>
  )
}
