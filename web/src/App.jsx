import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { createContext, useContext, useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { apiGet } from './lib/api'
import { ToastProvider } from './components/Toast'
import ErrorBoundary from './components/ErrorBoundary'
import LoginPage from './pages/LoginPage'
import MainPage from './pages/MainPage'
import AdminPage from './pages/AdminPage'
import AuditLogTab from './components/AuditLogTab'
import SettingsTab from './components/SettingsTab'
import StatusTab from './components/StatusTab'

const AuthContext = createContext(null)
const ThemeContext = createContext(null)

export function useAuth() {
  return useContext(AuthContext)
}

export function useTheme() {
  return useContext(ThemeContext)
}

function getInitialTheme() {
  const stored = localStorage.getItem('theme')
  if (stored === 'dark' || stored === 'light') return stored
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function LoadingScreen() {
  const { t } = useTranslation()
  return (
    <div className="flex items-center justify-center h-screen bg-base-200">
      <div className="flex flex-col items-center gap-3">
        <span className="loading loading-spinner text-primary"></span>
        <span className="text-sm text-base-content/50">{t('common.loading')}</span>
      </div>
    </div>
  )
}

export default function App() {
  const [config, setConfig] = useState(null)
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(true)
  const [theme, setTheme] = useState(getInitialTheme)

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme)
    document.documentElement.classList.toggle('dark', theme === 'dark')
    localStorage.setItem('theme', theme)
  }, [theme])

  const toggleTheme = useCallback(() => {
    setTheme(prev => prev === 'light' ? 'dark' : 'light')
  }, [])

  useEffect(() => {
    apiGet('/api/config')
      .then(data => {
        setConfig({ host: data.host, siteTitle: data.site_title, authMode: data.auth_mode })
        if (data.email) {
          setUser({ email: data.email })
        }
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    if (config?.siteTitle) {
      document.title = config.siteTitle
    }
  }, [config])

  if (loading) {
    return <LoadingScreen />
  }

  const isAuthenticated = !!user

  return (
    <ThemeContext.Provider value={{ theme, toggleTheme }}>
      <AuthContext.Provider value={{ config, user, isAuthenticated }}>
        <ToastProvider>
          <BrowserRouter>
            <ErrorBoundary>
              <Routes>
                <Route path="/login" element={isAuthenticated ? <Navigate to="/" /> : <LoginPage />} />
                <Route path="/" element={isAuthenticated ? <MainPage /> : <Navigate to="/login" />} />
                <Route path="/admin" element={isAuthenticated ? <AdminPage /> : <Navigate to="/login" />}>
                  <Route index element={<Navigate to="audit" replace />} />
                  <Route path="audit" element={<AuditLogTab />} />
                  <Route path="settings" element={<SettingsTab />} />
                  <Route path="status" element={<StatusTab />} />
                </Route>
                <Route path="*" element={<Navigate to="/" />} />
              </Routes>
            </ErrorBoundary>
          </BrowserRouter>
        </ToastProvider>
      </AuthContext.Provider>
    </ThemeContext.Provider>
  )
}
