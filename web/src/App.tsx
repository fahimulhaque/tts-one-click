import { Routes, Route, Link } from 'react-router-dom'
import { AppProvider } from './context/AppContext'
import StatusBar from './components/StatusBar'
import TTSPage from './pages/TTSPage'
import ClonePage from './pages/ClonePage'
import MetricsPage from './pages/MetricsPage'

export default function App() {
  return (
    <AppProvider>
      <div className="min-h-screen bg-gray-50 flex flex-col">
        <StatusBar />
        <nav className="bg-white border-b px-6 py-2 flex gap-6 text-sm font-medium">
          <Link to="/" className="text-blue-600 hover:text-blue-800">TTS</Link>
          <Link to="/clone" className="text-blue-600 hover:text-blue-800">Voice Clone</Link>
          <Link to="/metrics" className="text-blue-600 hover:text-blue-800">Metrics</Link>
        </nav>
        <main className="flex-1 p-6">
          <Routes>
            <Route path="/" element={<TTSPage />} />
            <Route path="/clone" element={<ClonePage />} />
            <Route path="/metrics" element={<MetricsPage />} />
          </Routes>
        </main>
      </div>
    </AppProvider>
  )
}
