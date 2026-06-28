import { useState, useRef } from 'react'
import { Upload, Loader2, Copy } from 'lucide-react'

interface Props { onAudio: (blob: Blob) => void }

export default function CloneForm({ onAudio }: Props) {
  const [file, setFile] = useState<File | null>(null)
  const [text, setText] = useState('')
  const [transcript, setTranscript] = useState('')
  const [loading, setLoading] = useState(false)
  const fileRef = useRef<HTMLInputElement>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!file) return
    setLoading(true)
    const fd = new FormData()
    fd.append('audio', file)
    fd.append('text', text)
    fd.append('transcript', transcript)
    const resp = await fetch('/api/v1/clone', { method: 'POST', body: fd })
    const blob = await resp.blob()
    setLoading(false)
    onAudio(blob)
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div
        onClick={() => fileRef.current?.click()}
        className="border-2 border-dashed border-gray-300 hover:border-blue-400 rounded-lg p-8 text-center cursor-pointer">
        <Upload className="h-8 w-8 mx-auto text-gray-400 mb-2" />
        <p className="font-medium text-gray-600">Upload reference audio</p>
        <p className="text-xs text-gray-400 mt-1">3–10 seconds, WAV/MP3</p>
        {file && <p className="text-sm text-green-600 mt-2">{file.name}</p>}
        <input ref={fileRef} type="file" accept="audio/*" className="hidden"
          onChange={e => setFile(e.target.files?.[0] ?? null)} />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Text to synthesize
        </label>
        <textarea className="w-full h-24 p-3 border rounded-lg resize-none"
          value={text} onChange={e => setText(e.target.value)} />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Reference audio transcript (optional)
        </label>
        <input type="text" className="w-full p-2 border rounded-lg"
          value={transcript} onChange={e => setTranscript(e.target.value)} />
      </div>
      <button type="submit" disabled={!file || !text || loading}
        className="w-full bg-purple-600 hover:bg-purple-700 disabled:opacity-50 text-white py-2 px-4 rounded-lg flex items-center justify-center gap-2">
        {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Copy className="h-4 w-4" />}
        {loading ? 'Cloning...' : 'Clone Voice'}
      </button>
    </form>
  )
}
