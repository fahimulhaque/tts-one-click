import { useQuery } from '@tanstack/react-query'

interface Metrics { requests: number; ttft_ms: number; rtf: number }

export function useMetrics() {
  return useQuery<Metrics>({
    queryKey: ['metrics'],
    queryFn: async () => {
      const r = await fetch('/api/v1/metrics')
      return r.json()
    },
    refetchInterval: 2000,
  })
}
