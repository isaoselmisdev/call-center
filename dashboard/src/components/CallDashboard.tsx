import { useState, useEffect, useRef } from 'react'
import { LogOut, Phone, Activity, CheckCircle2, Clock } from 'lucide-react'
import CallCard from './CallCard'

interface Agent {
  id: string
  name: string
  token: string
}

interface Call {
  call_id: string
  customer_number: string
  timestamp: string
  assigned_agent_id: string
  status: string
}

interface CallDashboardProps {
  agent: Agent
  onLogout: () => void
}

export default function CallDashboard({ agent, onLogout }: CallDashboardProps) {
  const [calls, setCalls] = useState<Call[]>([])
  const [loading, setLoading] = useState(true)
  const [wsConnected, setWsConnected] = useState(false)
  const wsRef = useRef<WebSocket | null>(null)

  useEffect(() => {
    // Fetch existing calls
    fetchCalls()

    // Connect to WebSocket with auth token
    const ws = new WebSocket(`ws://localhost:8082/ws/assigned?token=${encodeURIComponent(agent.token)}`)
    
    ws.onopen = () => {
      console.log('WebSocket connected')
      setWsConnected(true)
    }

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        if (data.type === 'new_call' && data.data) {
          setCalls((prev) => [data.data, ...prev])
          // Play notification sound
          new Audio('data:audio/wav;base64,UklGRnoGAABXQVZFZm10IBAAAAABAAEAQB8AAEAfAAABAAgAZGF0YQoGAACBhYqFbF1fdJivrJBhNjVgodDbq2EcBj+a2/LDciUFLIHO8tiJNwgZaLvt559NEAxQp+PwtmMcBjiR1/LMeSwFJHfH8N2QQAoUXrTp66hVFApGn+DyvmwhBSuBzvLbiTYIGWe77+ekUxAKUp/h8bJoGAY6k9bzy3osBip+zPPZeDAFLnvM8+OIQfM=').play()
        }
      } catch (err) {
        console.error('Failed to parse WebSocket message:', err)
      }
    }

    ws.onerror = (error) => {
      console.error('WebSocket error:', error)
      setWsConnected(false)
    }

    ws.onclose = () => {
      console.log('WebSocket disconnected')
      setWsConnected(false)
      // Attempt to reconnect after 5 seconds
      setTimeout(() => {
        if (wsRef.current?.readyState === WebSocket.CLOSED) {
          location.reload()
        }
      }, 5000)
    }

    wsRef.current = ws

    return () => {
      ws.close()
    }
  }, [agent.token])

  const fetchCalls = async () => {
    try {
      const response = await fetch('http://localhost:8082/api/v1/calls', {
        headers: {
          'Authorization': `Bearer ${agent.token}`,
        },
      })
      const data = await response.json()
      if (data.success && data.data) {
        setCalls(data.data)
      }
      setLoading(false)
    } catch (err) {
      console.error('Failed to fetch calls:', err)
      setLoading(false)
    }
  }

  const completeCall = async (callId: string) => {
    try {
      await fetch(`http://localhost:8082/api/v1/calls/${callId}/complete`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${agent.token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          status: 'completed',
          notes: 'Call completed successfully',
        }),
      })

      // Update local state
      setCalls((prev) => prev.map((call) => call.call_id === callId ? { ...call, status: 'completed' } : call))
    } catch (err) {
      console.error('Failed to complete call:', err)
    }
  }

  const pendingCalls = calls.filter((call) => call.status === 'assigned')
  const completedCalls = calls.filter((call) => call.status === 'completed')

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="border-b border-gray-200 bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center space-x-3">
              <div className="h-8 w-8 bg-black rounded-md flex items-center justify-center">
                <Phone className="h-4 w-4 text-white" />
              </div>
              <div>
                <h1 className="text-lg font-semibold text-gray-900">Call Center Dashboard</h1>
                <p className="text-xs text-gray-500">Agent: <span className="font-mono text-gray-900">{agent.id}</span></p>
              </div>
            </div>

            <div className="flex items-center space-x-4">
              <div className="flex items-center space-x-2">
                <div className={`h-2 w-2 rounded-full ${wsConnected ? 'bg-green-500' : 'bg-gray-300'}`} />
                <span className="text-xs text-gray-500">
                  {wsConnected ? 'Connected' : 'Disconnected'}
                </span>
              </div>
              
              <button
                onClick={onLogout}
                className="inline-flex items-center justify-center rounded-md text-sm font-medium border border-gray-300 bg-white px-3 py-2 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-gray-900 focus:ring-offset-2 transition-colors"
              >
                <LogOut className="h-4 w-4 mr-2" />
                Logout
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Stats */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <div className="bg-white rounded-lg border border-gray-200 p-6">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <Activity className="h-6 w-6 text-gray-900" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-500">Total Calls</p>
                <p className="text-2xl font-semibold text-gray-900">{calls.length}</p>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg border border-gray-200 p-6">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <Clock className="h-6 w-6 text-gray-900" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-500">Pending</p>
                <p className="text-2xl font-semibold text-gray-900">{pendingCalls.length}</p>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg border border-gray-200 p-6">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <CheckCircle2 className="h-6 w-6 text-gray-900" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-500">Completed</p>
                <p className="text-2xl font-semibold text-gray-900">{completedCalls.length}</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Calls List */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pb-8">
        <div className="mb-6">
          <h2 className="text-lg font-semibold text-gray-900">Active Calls</h2>
          <p className="text-sm text-gray-500">Manage and complete your assigned calls</p>
        </div>

        {loading ? (
          <div className="bg-white rounded-lg border border-gray-200 p-8 text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto mb-4"></div>
            <p className="text-sm text-gray-500">Loading calls...</p>
          </div>
        ) : calls.length === 0 ? (
          <div className="bg-white rounded-lg border border-gray-200 p-8 text-center">
            <Phone className="h-12 w-12 text-gray-300 mx-auto mb-4" />
            <p className="text-sm font-medium text-gray-900 mb-1">No calls yet</p>
            <p className="text-sm text-gray-500">Waiting for incoming calls...</p>
          </div>
        ) : (
          <div className="space-y-4">
            {calls.map((call) => (
              <CallCard
                key={call.call_id}
                call={call}
                onComplete={() => completeCall(call.call_id)}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
