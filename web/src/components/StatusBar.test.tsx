import { render, screen } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import StatusBar from './StatusBar'

vi.mock('../context/AppContext', () => ({
  useApp: () => ({ model: 'chatterbox', gpuAvailable: true, serverReady: true })
}))

describe('StatusBar', () => {
  it('shows model name', () => {
    render(<StatusBar />)
    expect(screen.getByText(/chatterbox/i)).toBeInTheDocument()
  })
  it('shows GPU badge when available', () => {
    render(<StatusBar />)
    expect(screen.getByText(/gpu/i)).toBeInTheDocument()
  })
})
