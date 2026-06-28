import { useRef, useCallback } from 'react'

interface WSTTSMessage {
  text: string
  model: string
  speed: number
  voice?: string
}

export function useWebSocket(onAudio: (blob: Blob) => void) {
  const wsRef = useRef<WebSocket | null>(null)
  const chunksRef = useRef<Uint8Array[]>([])

  const send = useCallback((msg: WSTTSMessage) => {
    chunksRef.current = []
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const ws = new WebSocket(`${protocol}//${window.location.host}/ws/tts`)
    wsRef.current = ws

    ws.onmessage = (e) => {
      if (e.data instanceof Blob) {
        e.data.arrayBuffer().then(buf => {
          chunksRef.current.push(new Uint8Array(buf))
        })
      }
    }
    ws.onclose = () => {
      if (chunksRef.current.length > 0) {
        const total = chunksRef.current.reduce((a, b) => a + b.length, 0)
        const merged = new Uint8Array(total)
        let offset = 0
        for (const chunk of chunksRef.current) {
          merged.set(chunk, offset)
          offset += chunk.length
        }
        onAudio(new Blob([merged], { type: 'audio/wav' }))
      }
    }
    ws.onopen = () => ws.send(JSON.stringify(msg))
  }, [onAudio])

  const cancel = useCallback(() => wsRef.current?.close(), [])
  return { send, cancel }
}
