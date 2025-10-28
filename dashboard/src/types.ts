export interface Agent {
  id: string
  name: string
  token: string
}

export interface Admin {
  username: string
  token: string
}

export interface User {
  type: 'agent' | 'admin'
  data: Agent | Admin
}

export interface Call {
  call_id: string
  customer_number: string
  timestamp: string
  assigned_agent_id: string
  status: string
}

export interface CreateAgentRequest {
  agent_id: string
  agent_name: string
  password: string
}
