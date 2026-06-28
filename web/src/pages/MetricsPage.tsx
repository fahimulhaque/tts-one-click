import { useMetrics } from '../hooks/useMetrics'
import { Activity, Clock, Zap } from 'lucide-react'

function StatCard({ icon: Icon, label, value, unit }: {
  icon: React.ElementType; label: string; value: string | number; unit: string
}) {
  return (
    <div className="bg-white rounded-xl border shadow-sm p-6 flex items-center gap-4">
      <div className="bg-blue-50 p-3 rounded-lg">
        <Icon className="h-6 w-6 text-blue-600" />
      </div>
      <div>
        <p className="text-sm text-gray-500">{label}</p>
        <p className="text-2xl font-bold text-gray-900">{value}<span className="text-sm font-normal text-gray-400 ml-1">{unit}</span></p>
      </div>
    </div>
  )
}

export default function MetricsPage() {
  const { data } = useMetrics()
  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Performance Metrics</h1>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <StatCard icon={Activity} label="Total Requests" value={data?.requests ?? 0} unit="req" />
        <StatCard icon={Clock} label="Avg TTFT" value={(data?.ttft_ms ?? 0).toFixed(0)} unit="ms" />
        <StatCard icon={Zap} label="Avg RTF" value={(data?.rtf ?? 0).toFixed(3)} unit="" />
      </div>
      <p className="text-xs text-gray-400 mt-4 text-center">Refreshes every 2 seconds</p>
    </div>
  )
}
