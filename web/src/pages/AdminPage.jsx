import { NavLink, Navigate, Outlet } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { ClipboardList, Settings2, Activity } from 'lucide-react'
import Navbar from '../components/Navbar'

const TABS = [
  { key: 'audit', icon: ClipboardList },
  { key: 'settings', icon: Settings2 },
  { key: 'status', icon: Activity },
]

export default function AdminPage() {
  const { t } = useTranslation()

  return (
    <div className="min-h-screen bg-base-200">
      <Navbar />
      <div className="max-w-7xl mx-auto px-4 sm:px-6 py-6">
        <h1 className="text-xl font-bold text-base-content mb-4">{t('admin.pageTitle')}</h1>
        <div className="flex items-center gap-1 border-b border-base-300/60 mb-6">
          {TABS.map(({ key, icon: Icon }) => (
            <NavLink
              key={key}
              to={`/admin/${key}`}
              className={({ isActive }) =>
                `px-4 py-2.5 text-sm font-medium border-b-2 -mb-px transition-colors duration-150 inline-flex items-center gap-1.5 ${
                  isActive
                    ? 'border-primary text-primary'
                    : 'border-transparent text-base-content/50 hover:text-base-content/70 hover:border-base-300'
                }`
              }
            >
              <Icon size={15} />
              {t(`admin.${key === 'audit' ? 'auditLogs' : key}`)}
            </NavLink>
          ))}
        </div>
        <Outlet />
      </div>
    </div>
  )
}
