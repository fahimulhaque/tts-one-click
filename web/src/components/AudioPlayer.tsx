import { useEffect, useRef } from 'react'

interface Props { blob: Blob | null }

export default function AudioPlayer({ blob }: Props) {
  const audioRef = useRef<HTMLAudioElement>(null)
  const urlRef = useRef<string | null>(null)

  useEffect(() => {
    if (!blob || !audioRef.current) return
    if (urlRef.current) URL.revokeObjectURL(urlRef.current)
    urlRef.current = URL.createObjectURL(blob)
    audioRef.current.src = urlRef.current
    audioRef.current.play()
    return () => { if (urlRef.current) URL.revokeObjectURL(urlRef.current) }
  }, [blob])

  if (!blob) return null

  return (
    <div className="mt-4 p-4 bg-gray-50 rounded-lg">
      <audio ref={audioRef} controls className="w-full" />
    </div>
  )
}
