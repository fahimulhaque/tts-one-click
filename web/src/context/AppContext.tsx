import { createContext, useContext, useState, useEffect, ReactNode } from 'react'

interface AppState {
  model: string
  gpuAvailable: boolean
  serverReady: boolean
}

const AppContext = createContext<AppState>({ model: '', gpuAvailable: false, serverReady: false })

export function AppProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AppState>({ model: '', gpuAvailable: false, serverReady: false })

  useEffect(() => {
    const poll = async () => {
      try {
        const r = await fetch('/api/v1/health')
        if (r.ok) {
          const d = await r.json()
          setState({ model: d.model, gpuAvailable: d.gpu, serverReady: true })
        }
      } catch { /* server not yet ready */ }
    }
    poll()
    const id = setInterval(poll, 10000)
    return () => clearInterval(id)
  }, [])

  return <AppContext.Provider value={state}>{children}</AppContext.Provider>
}

export const useApp = () => useContext(AppContext)
