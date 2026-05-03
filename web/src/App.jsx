import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { createContext, useContext, useState, useEffect, useCallback } from 'react'
import { apiGet } from './lib/api'
import LoginPage from './pages/LoginPage'
import MainPage from './pages/MainPage'
import AdminPage from './pages/AdminPage'

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
        setConfig({ host: data.host, siteTitle: data.site_title })
        if (data.email) {
          setUser({ email: data.email })
        }
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return (
      <div className="flex items-center justify-center h-screen bg-base-200">
        <div className="flex flex-col items-center gap-3">
          <span className="loading loading-spinner text-primary"></span>
          <span className="text-sm text-base-content/50">Loading...</span>
        </div>
      </div>
    )
  }

  const isAuthenticated = !!user

  return (
    <ThemeContext.Provider value={{ theme, toggleTheme }}>
      <AuthContext.Provider value={{ config, user, isAuthenticated }}>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={isAuthenticated ? <Navigate to="/" /> : <LoginPage />} />
            <Route path="/" element={isAuthenticated ? <MainPage /> : <Navigate to="/login" />} />
            <Route path="/admin" element={isAuthenticated ? <AdminPage /> : <Navigate to="/login" />} />
            <Route path="*" element={<Navigate to="/" />} />
          </Routes>
        </BrowserRouter>
      </AuthContext.Provider>
    </ThemeContext.Provider>
  )
}
