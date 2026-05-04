import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuth, useTheme } from '../App'
import { Sun, Moon, Shield, LogOut, Mail, Menu } from 'lucide-react'

export default function Navbar() {
  const { config, user } = useAuth()
  const { theme, toggleTheme } = useTheme()
  const { t, i18n } = useTranslation()

  return (
    <div className="sticky top-0 z-50 glass border-b border-base-300 backdrop-blur-xl">
      <div className="navbar max-w-7xl mx-auto px-4 h-14">
        <div className="flex-1">
          <Link to="/" className="flex items-center gap-2 text-lg font-semibold text-base-content hover:text-primary transition-colors">
            <Mail size={20} />
            {config?.siteTitle || 'Forsaken Mail'}
          </Link>
        </div>
        <div className="flex-none flex items-center gap-1">
          {/* Desktop: inline controls */}
          <div className="hidden sm:flex items-center gap-1">
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
              <span className="text-xs text-base-content/50 ml-1">{user.email}</span>
            )}

            <div className="w-px h-5 bg-base-300/60 mx-1"></div>

            <Link to="/admin" className="btn btn-ghost btn-sm gap-1" title={t('navbar.admin')}>
              <Shield size={16} />
              <span>{t('navbar.admin')}</span>
            </Link>
            <a href="/auth/logout" className="btn btn-ghost btn-sm btn-circle" title={t('navbar.logout')}>
              <LogOut size={16} />
            </a>
          </div>

          {/* Mobile: theme toggle */}
          <button
            className="btn btn-ghost btn-sm btn-circle sm:hidden"
            onClick={toggleTheme}
            title={theme === 'light' ? t('navbar.darkMode') : t('navbar.lightMode')}
          >
            {theme === 'light' ? <Moon size={16} /> : <Sun size={16} />}
          </button>

          {/* Mobile: dropdown menu */}
          <div className="dropdown dropdown-end sm:hidden">
            <div tabIndex={0} role="button" className="btn btn-ghost btn-sm btn-circle">
              <Menu size={18} />
            </div>
            <ul tabIndex={0} className="dropdown-content menu bg-base-100 rounded-box z-1 w-44 p-2 shadow-lg border border-base-300/60 mt-2">
              {/* Language */}
              <li className="menu-title text-xs">
                <span>{t('navbar.language')}</span>
              </li>
              <li>
                <button
                  className={i18n.language === 'en' ? 'active' : ''}
                  onClick={() => i18n.changeLanguage('en')}
                >
                  EN
                </button>
              </li>
              <li>
                <button
                  className={i18n.language === 'zh' ? 'active' : ''}
                  onClick={() => i18n.changeLanguage('zh')}
                >
                  中文
                </button>
              </li>
              <div className="divider my-0"></div>
              {/* Admin */}
              <li>
                <Link to="/admin">
                  <Shield size={14} />
                  {t('navbar.admin')}
                </Link>
              </li>
              {/* Logout */}
              <li>
                <a href="/auth/logout">
                  <LogOut size={14} />
                  {t('navbar.logout')}
                </a>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  )
}
