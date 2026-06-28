import { useState } from 'react'
import CloneForm from '../components/CloneForm'
import AudioPlayer from '../components/AudioPlayer'

export default function ClonePage() {
  const [audio, setAudio] = useState<Blob | null>(null)
  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Voice Cloning</h1>
      <div className="bg-white rounded-xl shadow-sm border p-6">
        <CloneForm onAudio={setAudio} />
        <AudioPlayer blob={audio} />
      </div>
    </div>
  )
}
