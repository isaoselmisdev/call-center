import { Phone, Clock, CheckCircle2, Circle } from 'lucide-react'

interface Call {
  call_id: string
  customer_number: string
  timestamp: string
  assigned_agent_id: string
  status: string
}

interface CallCardProps {
  call: Call
  onComplete: () => void
}

export default function CallCard({ call, onComplete }: CallCardProps) {
  const formattedTime = new Date(call.timestamp).toLocaleString()
  const isCompleted = call.status === 'completed'

  return (
    <div className={`group relative bg-white border border-gray-200 rounded-lg p-6 transition-all ${
      isCompleted ? 'opacity-60' : 'hover:border-gray-300 hover:shadow-sm'
    }`}>
      <div className="flex items-start justify-between">
        <div className="flex-1 space-y-3">
          {/* Status Badge */}
          <div className="flex items-center space-x-2">
            {isCompleted ? (
              <div className="inline-flex items-center rounded-md border border-gray-200 bg-gray-50 px-2.5 py-0.5 text-xs font-medium text-gray-900">
                <CheckCircle2 className="mr-1 h-3 w-3" />
                Completed
              </div>
            ) : (
              <div className="inline-flex items-center rounded-md border border-gray-900 bg-gray-900 px-2.5 py-0.5 text-xs font-medium text-white">
                <Circle className="mr-1 h-3 w-3 fill-white" />
                Active
              </div>
            )}
            <span className="text-xs font-mono text-gray-400">
              #{call.call_id.substring(0, 8)}
            </span>
          </div>

          {/* Customer Number */}
          <div className="flex items-center space-x-2">
            <Phone className="h-5 w-5 text-gray-400" />
            <span className="text-base font-semibold text-gray-900">
              {call.customer_number}
            </span>
          </div>

          {/* Timestamp */}
          <div className="flex items-center space-x-2">
            <Clock className="h-4 w-4 text-gray-400" />
            <span className="text-sm text-gray-500">{formattedTime}</span>
          </div>
        </div>

        {/* Complete Button */}
        {!isCompleted && (
          <button
            onClick={onComplete}
            className="inline-flex items-center justify-center rounded-md bg-black px-4 py-2 text-sm font-medium text-white hover:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-gray-900 focus:ring-offset-2 transition-colors"
          >
            <CheckCircle2 className="mr-2 h-4 w-4" />
            Complete
          </button>
        )}
      </div>
    </div>
  )
}
