# Call Center System

Modern event-driven call center platform with real-time agent management, built with Go, Kafka, PostgreSQL, Redis, and React.

## ✨ Features

- 🎯 **Real-time Call Distribution** - Round-robin assignment via Redis
- 👥 **Admin Dashboard** - Manage agents, view statistics, create new agents
- 📡 **Live Agent Sync** - Kafka-based real-time synchronization (no restarts needed)
- 🔔 **WebSocket Notifications** - Agents receive calls instantly
- 🔐 **JWT Authentication** - Secure admin and agent access
- 📊 **Kafka UI** - Monitor message flow with Kowl

## 🚀 Quick Start

```bash
# Clone and start all services
docker compose up -d

# Access the dashboard
open http://localhost:3000
```

**Services Running:**
- 🖥️ Dashboard: http://localhost:3000
- 📞 Call Center API: http://localhost:8081
- 👤 Agent API: http://localhost:8082
- 🔀 Distributor: http://localhost:8083
- 📊 Kafka UI: http://localhost:8080

**Default Admin Login:**
- Username: `admin`
- Password: `admin123`

## 🏗️ Architecture

```
Call Center API → Kafka (incoming_calls) → Distributor → Kafka (assigned_calls) → Agents
                                              ↕
                                           Redis (round-robin)
                                              ↕
                                           Kafka (agent_changes) ← Admin Dashboard
```

### Services

1. **Dashboard** (React + Vite) - Admin panel and agent portal
2. **Call Center API** - Receives incoming calls, publishes to Kafka
3. **Customer Agent API** - Authentication, call management, agent CRUD
4. **Distributor** - Assigns calls to agents via round-robin, syncs Redis

## 📋 Quick Tasks

### Create an Agent (via Dashboard)
1. Login at http://localhost:3000 as admin
2. Go to "Create Agent" tab
3. Fill in agent details → Submit
4. Agent instantly added to Redis (no restart needed!)

### Submit a Test Call
```bash
curl -X POST http://localhost:8081/api/v1/calls \
  -H "Content-Type: application/json" \
  -d '{"customer_number": "+1234567890"}'
```

### Login as Agent (via Dashboard)
1. Go to http://localhost:3000
2. Select "Agent Portal"
3. Login with agent credentials
4. See assigned calls in real-time

## 🔑 Key Concepts

### Real-time Agent Synchronization
When you create/delete an agent:
1. Agent saved to PostgreSQL
2. Event published to Kafka `agent_changes` topic
3. Distributor consumes event
4. Redis updated instantly
5. Agent ready to receive calls (no restart!)

### Kafka Topics
- `incoming_calls` - New customer calls
- `assigned_calls` - Calls assigned to agents
- `agent_changes` - Agent create/delete events

### API Authentication
- **Admin**: JWT with `agent_id="admin"`
- **Agents**: JWT with `agent_id=<agent_id>`
- All protected routes require `Authorization: Bearer <token>`

## 🛠️ Tech Stack

**Backend**: Go (Fiber), PostgreSQL, Redis, Kafka (KRaft)  
**Frontend**: React 19, TypeScript, Vite, TailwindCSS  
**Infrastructure**: Docker, Nginx

## 📁 Project Structure

```
├── cmd/                    # Service entry points
├── internal/               # Business logic
├── pkg/                    # Shared packages (database, auth, logger)
├── models/                 # Data models
├── dashboard/              # React frontend
└── docker compose.yml      # All services configuration
```

## 🔍 Monitoring & Debugging

```bash
# View logs
docker compose logs -f [service-name]

# Check Redis agents
docker exec callcenter-redis redis-cli LRANGE available_agents 0 -1

# Kafka UI
open http://localhost:8080

# Service health
curl http://localhost:8081/health
curl http://localhost:8082/health
curl http://localhost:8083/health
```




