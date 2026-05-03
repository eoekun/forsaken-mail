import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuth, useTheme } from '../App'
import { Sun, Moon, Shield, LogOut } from 'lucide-react'

export default function Navbar() {
  const { config, user } = useAuth()
  const { theme, toggleTheme } = useTheme()
  const { t, i18n } = useTranslation()

  return (
    <div className="sticky top-0 z-50 glass border-b border-base-300/40">
      <div className="navbar max-w-7xl mx-auto px-4 h-14">
        <div className="flex-1">
          <Link to="/" className="text-lg font-semibold text-base-content hover:text-primary transition-colors">
            {config?.siteTitle || 'Forsaken Mail'}
          </Link>
        </div>
        <div className="flex-none flex items-center gap-1">
          <div className="join">
            <button
              className={`join-item btn btn-xs btn-ghost ${i18n.language === 'en' ? 'btn-active' : ''}`}
              onClick={() => i18n.changeLanguage('en')}
            >
              EN
            </button>
            <button
              className={`join-item btn btn-xs btn-ghost ${i18n.language === 'zh' ? 'btn-active' : ''}`}
              onClick={() => i18n.changeLanguage('zh')}
            >
              中
            </button>
          </div>

          <div className="w-px h-5 bg-base-300/60 mx-1"></div>

          <button
            className="btn btn-ghost btn-sm btn-circle"
            onClick={toggleTheme}
            title={theme === 'light' ? t('navbar.darkMode') : t('navbar.lightMode')}
          >
            {theme === 'light' ? <Moon size={16} /> : <Sun size={16} />}
          </button>

          {user && (
            <span className="text-xs text-base-content/50 ml-1 hidden sm:inline">{user.email}</span>
          )}

          <div className="w-px h-5 bg-base-300/60 mx-1 hidden sm:block"></div>

          <Link to="/admin" className="btn btn-ghost btn-sm btn-circle" title={t('navbar.admin')}>
            <Shield size={16} />
          </Link>
          <a href="/auth/logout" className="btn btn-ghost btn-sm btn-circle" title={t('navbar.logout')}>
            <LogOut size={16} />
          </a>
        </div>
      </div>
    </div>
  )
}
