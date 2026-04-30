import { Link } from 'react-router-dom'
import { useAuth } from '../App'

export default function Navbar() {
  const { config, user } = useAuth()

  return (
    <div className="navbar bg-base-100 shadow-md">
      <div className="flex-1">
        <Link to="/" className="btn btn-ghost text-xl">{config?.siteTitle || 'Forsaken Mail'}</Link>
      </div>
      <div className="flex-none gap-2">
        {user && <span className="text-sm text-base-content/70">{user.email}</span>}
        <Link to="/admin" className="btn btn-sm btn-ghost">Admin</Link>
        <a href="/auth/logout" className="btn btn-sm btn-ghost">Logout</a>
      </div>
    </div>
  )
}
