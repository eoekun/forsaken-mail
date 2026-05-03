import { useTranslation } from 'react-i18next'
import { useAuth } from '../App'
import { Mail, Github } from 'lucide-react'

export default function LoginPage() {
  const { config } = useAuth()
  const { t } = useTranslation()
  const params = new URLSearchParams(window.location.search)
  const error = params.get('error')

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-base-200 via-base-200 to-primary/5 px-4">
      <div className="w-full max-w-sm">
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-14 h-14 rounded-2xl bg-primary/10 mb-4">
            <Mail className="text-primary" size={28} />
          </div>
          <h1 className="text-2xl font-bold text-base-content">
            {config?.siteTitle || t('login.title')}
          </h1>
          <p className="text-sm text-base-content/50 mt-1">
            {t('login.subtitle')}
          </p>
        </div>

        {error === 'unauthorized_email' && (
          <div className="alert alert-error mb-6 rounded-lg text-sm">
            <span>{t('login.unauthorized')}</span>
          </div>
        )}

        <div className="card-modern p-6">
          <div className="flex flex-col gap-3">
            <a
              href="/auth/github/login"
              className="btn-modern btn-outline justify-center gap-3 h-11"
            >
              <Github size={18} />
              {t('login.github')}
            </a>
            <a
              href="/auth/google/login"
              className="btn-modern btn-outline justify-center gap-3 h-11"
            >
              <svg className="w-[18px] h-[18px]" viewBox="0 0 24 24">
                <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 01-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z"/>
                <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
                <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
                <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
              </svg>
              {t('login.google')}
            </a>
          </div>
        </div>

        <p className="text-center text-xs text-base-content/30 mt-6">
          Disposable email service
        </p>
      </div>
    </div>
  )
}
