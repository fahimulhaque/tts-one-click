import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import TTSForm from './TTSForm'

vi.mock('../context/AppContext', () => ({
  useApp: () => ({ model: 'chatterbox', gpuAvailable: true, serverReady: true })
}))

vi.mock('../hooks/useWebSocket', () => ({
  useWebSocket: () => ({ send: vi.fn(), cancel: vi.fn() })
}))

describe('TTSForm', () => {
  it('disables submit when text is empty', () => {
    render(<TTSForm onAudio={vi.fn()} />)
    const btn = screen.getByRole('button', { name: /generate/i })
    expect(btn).toBeDisabled()
  })

  it('shows character count', () => {
    render(<TTSForm onAudio={vi.fn()} />)
    const ta = screen.getByPlaceholderText(/enter text/i)
    fireEvent.change(ta, { target: { value: 'hello' } })
    expect(screen.getByText(/5 \/ 500/)).toBeInTheDocument()
  })
})
