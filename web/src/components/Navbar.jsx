import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuth } from '../App'

export default function Navbar() {
  const { config, user } = useAuth()
  const { t, i18n } = useTranslation()

  const handleLangChange = (e) => {
    i18n.changeLanguage(e.target.value)
  }

  return (
    <div className="navbar bg-base-100 shadow-md">
      <div className="flex-1">
        <Link to="/" className="btn btn-ghost text-xl">{config?.siteTitle || 'Forsaken Mail'}</Link>
      </div>
      <div className="flex-none gap-2">
        <select
          className="select select-sm select-ghost"
          value={i18n.language}
          onChange={handleLangChange}
          title={t('navbar.language')}
        >
          <option value="en">English</option>
          <option value="zh">中文</option>
        </select>
        {user && <span className="text-sm text-base-content/70">{user.email}</span>}
        <Link to="/admin" className="btn btn-sm btn-ghost">{t('navbar.admin')}</Link>
        <a href="/auth/logout" className="btn btn-sm btn-ghost">{t('navbar.logout')}</a>
      </div>
    </div>
  )
}
