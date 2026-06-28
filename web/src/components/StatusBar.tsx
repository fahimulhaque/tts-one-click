import { useApp } from '../context/AppContext'
import { Mic2, Cpu } from 'lucide-react'

export default function StatusBar() {
  const { model, gpuAvailable, serverReady } = useApp()
  return (
    <header className="bg-gray-900 text-white px-6 py-3 flex items-center justify-between">
      <div className="flex items-center gap-2">
        <Mic2 className="h-5 w-5 text-blue-400" />
        <span className="font-semibold text-lg">TTS One-Click</span>
      </div>
      <div className="flex items-center gap-3 text-sm">
        {model && (
          <span className="bg-blue-600 px-2 py-0.5 rounded capitalize">{model}</span>
        )}
        {gpuAvailable && (
          <span className="bg-green-700 px-2 py-0.5 rounded flex items-center gap-1">
            <Cpu className="h-3 w-3" /> GPU
          </span>
        )}
        <span className={`h-2 w-2 rounded-full ${serverReady ? 'bg-green-400' : 'bg-red-400'}`} />
      </div>
    </header>
  )
}
