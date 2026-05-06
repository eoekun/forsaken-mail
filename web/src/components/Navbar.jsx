import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuth, useTheme } from '../App'
import { Sun, Moon, Shield, LogOut, Mail, Menu, Clock, KeyRound, Check } from 'lucide-react'
import { formatMailTime } from '../lib/formatTime'
import { formatSender } from '../lib/formatSender'
import { useCopyToClipboard } from '../hooks/useCopyToClipboard'

export default function Navbar({ recentMails, onOpenRecentMail }) {
  const { config, user } = useAuth()
  const { theme, toggleTheme } = useTheme()
  const { t, i18n } = useTranslation()
  const { copiedId, copy } = useCopyToClipboard()

  const handleCopyCode = (e, mail) => {
    e.stopPropagation()
    const code = mail.extracted_codes?.[0]
    if (!code) return
    copy(code, mail.id)
  }

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
            {/* Recent Mails dropdown */}
            {recentMails && recentMails.length > 0 && (
              <div className="dropdown dropdown-end">
                <div tabIndex={0} role="button" className="btn btn-xs btn-ghost gap-1">
                  <Clock size={13} />
                  <span className="text-xs">{recentMails.length}</span>
                </div>
                <div tabIndex={0} className="dropdown-content bg-base-100 rounded-xl shadow-lg border border-base-300/60 mt-2 w-80 max-h-80 overflow-y-auto z-50">
                  <div className="px-3 py-2 text-xs font-medium text-base-content/50 border-b border-base-300/40">
                    {t('recentMails.title')}
                  </div>
                  {recentMails.map((mail) => {
                    const recipient = mail.short_id || (mail.to_addr || mail.to || '').split('@')[0]
                    const sender = mail.from_addr || mail.from || ''
                    const codes = mail.extracted_codes || []
                    return (
                      <div
                        key={mail.id}
                        className="px-3 py-2 flex items-center gap-2 cursor-pointer hover:bg-base-200 transition-colors border-b border-base-300/20 last:border-0"
                        onClick={() => {
                          onOpenRecentMail(mail)
                          document.activeElement?.blur()
                        }}
                      >
                        <div className="min-w-0 flex-1">
                          <div className="flex items-center gap-1.5 mb-0.5">
                            {recipient && (
                              <span className="text-[11px] font-mono text-primary/70 bg-primary/5 px-1 py-0.5 rounded shrink-0">
                                {recipient}
                              </span>
                            )}
                            <span className="text-[11px] text-base-content/50 truncate">{formatSender(sender)}</span>
                          </div>
                          <p className="text-xs text-base-content/60 truncate">{mail.subject || t('mailList.noSubject')}</p>
                        </div>
                        {codes.length > 0 && (
                          <button
                            onClick={(e) => handleCopyCode(e, mail)}
                            className={`shrink-0 inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[11px] font-mono font-semibold transition-all cursor-pointer ${
                              copiedId === mail.id
                                ? 'bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400'
                                : 'bg-primary/10 text-primary hover:bg-primary/20'
                            }`}
                          >
                            {copiedId === mail.id ? <Check size={10} /> : <KeyRound size={10} />}
                          </button>
                        )}
                        <span className="text-[11px] text-base-content/30 shrink-0 tabular-nums">
                          {formatMailTime(mail.created_at, t)}
                        </span>
                      </div>
                    )
                  })}
                </div>
              </div>
            )}

            <div className="w-px h-5 bg-base-300/60 mx-1"></div>

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
            <ul tabIndex={0} className="dropdown-content menu bg-base-100 rounded-box z-1 w-56 p-2 shadow-lg border border-base-300/60 mt-2 max-h-[70vh] overflow-y-auto">
              {/* Recent Mails */}
              {recentMails && recentMails.length > 0 && (
                <>
                  <li className="menu-title text-xs">
                    <span><Clock size={12} /> {t('recentMails.title')} ({recentMails.length})</span>
                  </li>
                  {recentMails.slice(0, 5).map((mail) => {
                    const recipient = mail.short_id || (mail.to_addr || mail.to || '').split('@')[0]
                    return (
                      <li key={mail.id}>
                        <button
                          className="text-xs py-2"
                          onClick={() => onOpenRecentMail(mail)}
                        >
                          <span className="font-mono text-primary/70 text-[11px]">{recipient}</span>
                          <span className="truncate text-base-content/50">{mail.subject || t('mailList.noSubject')}</span>
                        </button>
                      </li>
                    )
                  })}
                  <div className="divider my-0"></div>
                </>
              )}
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
