import { createContext, useContext, useState, useCallback, useRef, useEffect } from 'react'
import { createPortal } from 'react-dom'
import { CheckCircle, XCircle, Info, X } from 'lucide-react'

const ToastContext = createContext(null)

export function useToast() {
  return useContext(ToastContext)
}

let toastId = 0

function ToastItem({ toast, onRemove }) {
  const [exiting, setExiting] = useState(false)
  const timerRef = useRef(null)

  useEffect(() => {
    timerRef.current = setTimeout(() => {
      setExiting(true)
      setTimeout(() => onRemove(toast.id), 300)
    }, 3000)
    return () => clearTimeout(timerRef.current)
  }, [toast.id, onRemove])

  const handleClose = () => {
    clearTimeout(timerRef.current)
    setExiting(true)
    setTimeout(() => onRemove(toast.id), 300)
  }

  const icons = {
    success: <CheckCircle size={16} className="text-success shrink-0" />,
    error: <XCircle size={16} className="text-error shrink-0" />,
    info: <Info size={16} className="text-info shrink-0" />,
  }

  const borders = {
    success: 'border-l-3 border-success',
    error: 'border-l-3 border-error',
    info: 'border-l-3 border-info',
  }

  return (
    <div
      className={`flex items-center gap-2 px-4 py-3 rounded-lg bg-base-100 shadow-lg border border-base-300/60 ${borders[toast.type]} transition-all duration-300 ${
        exiting ? 'opacity-0 translate-x-4' : 'opacity-100 translate-x-0'
      }`}
    >
      {icons[toast.type]}
      <span className="text-sm text-base-content flex-1">{toast.message}</span>
      <button
        className="btn btn-xs btn-ghost btn-circle text-base-content/40 hover:text-base-content"
        onClick={handleClose}
      >
        <X size={12} />
      </button>
    </div>
  )
}

export function ToastProvider({ children }) {
  const [toasts, setToasts] = useState([])

  const removeToast = useCallback((id) => {
    setToasts(prev => prev.filter(t => t.id !== id))
  }, [])

  const addToast = useCallback((type, message) => {
    const id = ++toastId
    setToasts(prev => [...prev, { id, type, message }])
  }, [])

  const toast = useCallback({
    success: (msg) => addToast('success', msg),
    error: (msg) => addToast('error', msg),
    info: (msg) => addToast('info', msg),
  }, [addToast])

  return (
    <ToastContext.Provider value={toast}>
      {children}
      {createPortal(
        <div className="fixed top-16 right-4 z-[100] flex flex-col gap-2 w-72">
          {toasts.map(t => (
            <ToastItem key={t.id} toast={t} onRemove={removeToast} />
          ))}
        </div>,
        document.body
      )}
    </ToastContext.Provider>
  )
}
