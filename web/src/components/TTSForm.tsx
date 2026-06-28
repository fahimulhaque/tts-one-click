import { useState } from 'react'
import { useApp } from '../context/AppContext'
import { useWebSocket } from '../hooks/useWebSocket'
import { Loader2, Play } from 'lucide-react'

const CHATTERBOX_TAGS = ['[laugh]', '[chuckle]', '[cough]', '[sigh]', '[gasp]']

interface Props { onAudio: (blob: Blob) => void }

export default function TTSForm({ onAudio }: Props) {
  const { model } = useApp()
  const [text, setText] = useState('')
  const [speed, setSpeed] = useState(1.0)
  const [loading, setLoading] = useState(false)
  const { send } = useWebSocket((blob) => { setLoading(false); onAudio(blob) })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    send({ text, model: model || 'chatterbox', speed })
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <textarea
          className="w-full h-32 p-3 border rounded-lg resize-none focus:ring-2 focus:ring-blue-500"
          placeholder="Enter text to synthesize..."
          value={text}
          onChange={e => setText(e.target.value.slice(0, 500))}
        />
        <div className="text-right text-xs text-gray-400">{text.length} / 500</div>
      </div>

      {model === 'chatterbox' && (
        <div className="flex gap-2 flex-wrap">
          {CHATTERBOX_TAGS.map(tag => (
            <button key={tag} type="button"
              onClick={() => setText(t => t + ' ' + tag)}
              className="text-xs bg-gray-100 hover:bg-gray-200 px-2 py-1 rounded">
              {tag}
            </button>
          ))}
        </div>
      )}

      <div className="flex gap-4 items-center">
        <label className="text-sm font-medium text-gray-700">Speed</label>
        <input type="range" min={0.5} max={2.0} step={0.1} value={speed}
          onChange={e => setSpeed(Number(e.target.value))}
          className="flex-1" />
        <span className="text-sm w-8">{speed.toFixed(1)}x</span>
      </div>

      <button type="submit" disabled={!text || loading}
        className="w-full bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white py-2 px-4 rounded-lg flex items-center justify-center gap-2">
        {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Play className="h-4 w-4" />}
        {loading ? 'Generating...' : 'Generate'}
      </button>
    </form>
  )
}
