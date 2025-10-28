import { useState, useEffect } from 'react'
import { Shield, Phone } from 'lucide-react'
import Login from './components/Login'
import AdminLogin from './components/AdminLogin'
import CallDashboard from './components/CallDashboard'
import AdminDashboard from './components/AdminDashboard'
import type { Agent, Admin, User } from './types'

type ViewMode = 'select' | 'agent-login' | 'admin-login' | 'agent-dashboard' | 'admin-dashboard'

function App() {
  const [user, setUser] = useState<User | null>(null)
  const [viewMode, setViewMode] = useState<ViewMode>('select')

  useEffect(() => {
    // Check if user is already logged in
    const storedUser = localStorage.getItem('user')
    if (storedUser) {
      try {
        const parsedUser: User = JSON.parse(storedUser)
        setUser(parsedUser)
        setViewMode(parsedUser.type === 'admin' ? 'admin-dashboard' : 'agent-dashboard')
      } catch (e) {
        localStorage.removeItem('user')
      }
    }
  }, [])

  const handleAgentLogin = (agentData: Agent) => {
    const userData: User = { type: 'agent', data: agentData }
    setUser(userData)
    localStorage.setItem('user', JSON.stringify(userData))
    setViewMode('agent-dashboard')
  }

  const handleAdminLogin = (adminData: Admin) => {
    const userData: User = { type: 'admin', data: adminData }
    setUser(userData)
    localStorage.setItem('user', JSON.stringify(userData))
    setViewMode('admin-dashboard')
  }

  const handleLogout = () => {
    setUser(null)
    localStorage.removeItem('user')
    setViewMode('select')
  }

  // View selection screen
  if (viewMode === 'select') {
    return (
      <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 flex items-center justify-center p-4">
        <div className="w-full max-w-4xl">
          <div className="text-center mb-12">
            <h1 className="text-4xl font-bold text-gray-900 mb-4">Call Center Portal</h1>
            <p className="text-lg text-gray-600">Choose how you'd like to sign in</p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* Agent Login Card */}
            <button
              onClick={() => setViewMode('agent-login')}
              className="bg-white rounded-2xl shadow-lg border border-gray-200 p-8 hover:shadow-xl hover:scale-105 transition-all duration-200 text-left group"
            >
              <div className="flex items-center justify-center w-16 h-16 bg-gradient-to-br from-blue-500 to-blue-600 rounded-2xl mb-6 group-hover:scale-110 transition-transform">
                <Phone className="h-8 w-8 text-white" />
              </div>
              <h2 className="text-2xl font-bold text-gray-900 mb-3">Agent Portal</h2>
              <p className="text-gray-600 mb-4">
                Sign in to manage and handle customer calls assigned to you
              </p>
              <div className="inline-flex items-center text-blue-600 font-medium">
                Continue as Agent
                <svg className="ml-2 h-5 w-5 group-hover:translate-x-1 transition-transform" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                </svg>
              </div>
            </button>

            {/* Admin Login Card */}
            <button
              onClick={() => setViewMode('admin-login')}
              className="bg-white rounded-2xl shadow-lg border border-gray-200 p-8 hover:shadow-xl hover:scale-105 transition-all duration-200 text-left group"
            >
              <div className="flex items-center justify-center w-16 h-16 bg-gradient-to-br from-gray-800 to-gray-900 rounded-2xl mb-6 group-hover:scale-110 transition-transform">
                <Shield className="h-8 w-8 text-white" />
              </div>
              <h2 className="text-2xl font-bold text-gray-900 mb-3">Admin Portal</h2>
              <p className="text-gray-600 mb-4">
                Access administrative controls and manage call center operations
              </p>
              <div className="inline-flex items-center text-gray-900 font-medium">
                Continue as Admin
                <svg className="ml-2 h-5 w-5 group-hover:translate-x-1 transition-transform" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                </svg>
              </div>
            </button>
          </div>
        </div>
      </div>
    )
  }

  // Agent Login
  if (viewMode === 'agent-login') {
    return (
      <div>
        <button
          onClick={() => setViewMode('select')}
          className="absolute top-4 left-4 text-sm text-gray-600 hover:text-gray-900 flex items-center"
        >
          <svg className="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
          Back
        </button>
        <Login onLogin={handleAgentLogin} />
      </div>
    )
  }

  // Admin Login
  if (viewMode === 'admin-login') {
    return (
      <div>
        <button
          onClick={() => setViewMode('select')}
          className="absolute top-4 left-4 text-sm text-gray-600 hover:text-gray-900 flex items-center"
        >
          <svg className="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
          Back
        </button>
        <AdminLogin onLogin={handleAdminLogin} />
      </div>
    )
  }

  // Agent Dashboard
  if (viewMode === 'agent-dashboard' && user?.type === 'agent') {
    return <CallDashboard agent={user.data as Agent} onLogout={handleLogout} />
  }

  // Admin Dashboard
  if (viewMode === 'admin-dashboard' && user?.type === 'admin') {
    return <AdminDashboard admin={user.data as Admin} onLogout={handleLogout} />
  }

  // Fallback
  return null
}

export default App