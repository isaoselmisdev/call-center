import { useState, useEffect } from 'react'
import { LogOut, Shield, Users, Phone, Activity, Trash2 } from 'lucide-react'
import CreateAgentForm from './CreateAgentForm'
import type { Admin } from '../types'

interface AdminDashboardProps {
  admin: Admin
  onLogout: () => void
}

interface AgentStats {
  agent_id: string
  agent_name: string
  status: string
  total_calls: number
  completed_calls: number
}

interface SystemStats {
  total_agents: number
  active_agents: number
  total_calls: number
  pending_calls: number
}

export default function AdminDashboard({ admin, onLogout }: AdminDashboardProps) {
  const [agents, setAgents] = useState<AgentStats[]>([])
  const [stats, setStats] = useState<SystemStats>({
    total_agents: 0,
    active_agents: 0,
    total_calls: 0,
    pending_calls: 0,
  })
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState<'overview' | 'create'>('overview')
  const [deletingAgent, setDeletingAgent] = useState<string | null>(null)
  const [confirmDelete, setConfirmDelete] = useState<string | null>(null)

  useEffect(() => {
    fetchDashboardData()
    // Refresh data every 10 seconds
    const interval = setInterval(fetchDashboardData, 10000)
    return () => clearInterval(interval)
  }, [admin.token])

  const fetchDashboardData = async () => {
    try {
      // Fetch agents list (this endpoint might need to be created in the backend)
      const agentsResponse = await fetch('http://localhost:8082/api/v1/agents/stats', {
        headers: {
          'Authorization': `Bearer ${admin.token}`,
        },
      })

      if (agentsResponse.ok) {
        const agentsData = await agentsResponse.json()
        if (agentsData.success && agentsData.data) {
          setAgents(agentsData.data)
          
          // Calculate stats
          const totalAgents = agentsData.data.length
          const activeAgents = agentsData.data.filter((a: AgentStats) => a.status === 'active').length
          const totalCalls = agentsData.data.reduce((sum: number, a: AgentStats) => sum + a.total_calls, 0)
          const completedCalls = agentsData.data.reduce((sum: number, a: AgentStats) => sum + a.completed_calls, 0)
          
          setStats({
            total_agents: totalAgents,
            active_agents: activeAgents,
            total_calls: totalCalls,
            pending_calls: totalCalls - completedCalls,
          })
        }
      }
      
      setLoading(false)
    } catch (err) {
      console.error('Failed to fetch dashboard data:', err)
      setLoading(false)
    }
  }

  const handleCreateAgent = async () => {
    // Refresh the agents list after creating a new agent
    await fetchDashboardData()
    // Switch back to overview tab
    setTimeout(() => setActiveTab('overview'), 1000)
  }

  const handleDeleteAgent = async (agentId: string) => {
    if (deletingAgent) return // Prevent multiple deletes at once

    setDeletingAgent(agentId)
    
    try {
      const response = await fetch(`http://localhost:8082/api/v1/agents/${agentId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${admin.token}`,
        },
      })

      const data = await response.json()

      if (response.ok && data.success) {
        // Refresh the agents list
        await fetchDashboardData()
        setConfirmDelete(null)
      } else {
        alert(data.message || 'Failed to delete agent')
      }
    } catch (err) {
      console.error('Failed to delete agent:', err)
      alert('Failed to delete agent. Please try again.')
    } finally {
      setDeletingAgent(null)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="border-b border-gray-200 bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center space-x-3">
              <div className="h-8 w-8 bg-gray-900 rounded-md flex items-center justify-center">
                <Shield className="h-4 w-4 text-white" />
              </div>
              <div>
                <h1 className="text-lg font-semibold text-gray-900">Admin Dashboard</h1>
                <p className="text-xs text-gray-500">
                  Logged in as: <span className="font-mono text-gray-900">{admin.username}</span>
                </p>
              </div>
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
      </header>

      {/* Tabs */}
      <div className="border-b border-gray-200 bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <nav className="flex space-x-8">
            <button
              onClick={() => setActiveTab('overview')}
              className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
                activeTab === 'overview'
                  ? 'border-gray-900 text-gray-900'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              <Activity className="h-4 w-4 inline-block mr-2" />
              Overview
            </button>
            <button
              onClick={() => setActiveTab('create')}
              className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
                activeTab === 'create'
                  ? 'border-gray-900 text-gray-900'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              <Users className="h-4 w-4 inline-block mr-2" />
              Create Agent
            </button>
          </nav>
        </div>
      </div>

      {/* Content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {activeTab === 'overview' ? (
          <>
            {/* Stats Grid */}
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4 mb-8">
              <div className="bg-white rounded-lg border border-gray-200 p-6">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <Users className="h-6 w-6 text-gray-900" />
                  </div>
                  <div className="ml-4">
                    <p className="text-sm font-medium text-gray-500">Total Agents</p>
                    <p className="text-2xl font-semibold text-gray-900">{stats.total_agents}</p>
                  </div>
                </div>
              </div>

              <div className="bg-white rounded-lg border border-gray-200 p-6">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <Activity className="h-6 w-6 text-green-600" />
                  </div>
                  <div className="ml-4">
                    <p className="text-sm font-medium text-gray-500">Active Agents</p>
                    <p className="text-2xl font-semibold text-gray-900">{stats.active_agents}</p>
                  </div>
                </div>
              </div>

              <div className="bg-white rounded-lg border border-gray-200 p-6">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <Phone className="h-6 w-6 text-gray-900" />
                  </div>
                  <div className="ml-4">
                    <p className="text-sm font-medium text-gray-500">Total Calls</p>
                    <p className="text-2xl font-semibold text-gray-900">{stats.total_calls}</p>
                  </div>
                </div>
              </div>

              <div className="bg-white rounded-lg border border-gray-200 p-6">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <Phone className="h-6 w-6 text-orange-600" />
                  </div>
                  <div className="ml-4">
                    <p className="text-sm font-medium text-gray-500">Pending Calls</p>
                    <p className="text-2xl font-semibold text-gray-900">{stats.pending_calls}</p>
                  </div>
                </div>
              </div>
            </div>

            {/* Agents Table */}
            <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
              <div className="px-6 py-4 border-b border-gray-200">
                <h2 className="text-lg font-semibold text-gray-900">Agents Overview</h2>
                <p className="text-sm text-gray-500">Manage and monitor all call center agents</p>
              </div>

              {loading ? (
                <div className="p-8 text-center">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto mb-4"></div>
                  <p className="text-sm text-gray-500">Loading agents...</p>
                </div>
              ) : agents.length === 0 ? (
                <div className="p-8 text-center">
                  <Users className="h-12 w-12 text-gray-300 mx-auto mb-4" />
                  <p className="text-sm font-medium text-gray-900 mb-1">No agents yet</p>
                  <p className="text-sm text-gray-500">Create your first agent to get started</p>
                  <button
                    onClick={() => setActiveTab('create')}
                    className="mt-4 inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-gray-900 hover:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-900"
                  >
                    Create Agent
                  </button>
                </div>
              ) : (
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Agent ID
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Name
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Status
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Total Calls
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Completed
                        </th>
                        <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Actions
                        </th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                      {agents.map((agent) => (
                        <tr key={agent.agent_id} className="hover:bg-gray-50">
                          <td className="px-6 py-4 whitespace-nowrap">
                            <div className="text-sm font-mono text-gray-900">{agent.agent_id}</div>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <div className="text-sm text-gray-900">{agent.agent_name}</div>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <span
                              className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                                agent.status === 'active'
                                  ? 'bg-green-100 text-green-800'
                                  : 'bg-gray-100 text-gray-800'
                              }`}
                            >
                              {agent.status}
                            </span>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {agent.total_calls}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {agent.completed_calls}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                            {confirmDelete === agent.agent_id ? (
                              <div className="flex items-center justify-end space-x-2">
                                <span className="text-xs text-gray-500">Delete?</span>
                                <button
                                  onClick={() => handleDeleteAgent(agent.agent_id)}
                                  disabled={deletingAgent === agent.agent_id}
                                  className="text-red-600 hover:text-red-900 font-medium disabled:opacity-50"
                                >
                                  {deletingAgent === agent.agent_id ? 'Deleting...' : 'Yes'}
                                </button>
                                <button
                                  onClick={() => setConfirmDelete(null)}
                                  disabled={deletingAgent === agent.agent_id}
                                  className="text-gray-600 hover:text-gray-900 font-medium disabled:opacity-50"
                                >
                                  No
                                </button>
                              </div>
                            ) : (
                              <button
                                onClick={() => setConfirmDelete(agent.agent_id)}
                                disabled={agent.status === 'inactive'}
                                className="text-red-600 hover:text-red-900 disabled:opacity-40 disabled:cursor-not-allowed inline-flex items-center"
                                title={agent.status === 'inactive' ? 'Agent already inactive' : 'Delete agent'}
                              >
                                <Trash2 className="h-4 w-4 mr-1" />
                                Delete
                              </button>
                            )}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          </>
        ) : (
          <div className="max-w-2xl">
            <CreateAgentForm onCreateAgent={handleCreateAgent} token={admin.token} />
          </div>
        )}
      </div>
    </div>
  )
}

