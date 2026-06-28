import { useState } from 'react'
import TTSForm from '../components/TTSForm'
import AudioPlayer from '../components/AudioPlayer'

export default function TTSPage() {
  const [audio, setAudio] = useState<Blob | null>(null)
  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Text to Speech</h1>
      <div className="bg-white rounded-xl shadow-sm border p-6">
        <TTSForm onAudio={setAudio} />
        <AudioPlayer blob={audio} />
      </div>
    </div>
  )
}
