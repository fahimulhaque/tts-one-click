import { render, screen } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import CloneForm from './CloneForm'

describe('CloneForm', () => {
  it('shows file upload area', () => {
    render(<CloneForm onAudio={vi.fn()} />)
    expect(screen.getByText(/upload reference audio/i)).toBeInTheDocument()
  })
  it('disables submit without file', () => {
    render(<CloneForm onAudio={vi.fn()} />)
    expect(screen.getByRole('button', { name: /clone/i })).toBeDisabled()
  })
})
