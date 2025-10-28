# Call Center Agent Dashboard

A modern React dashboard for call center agents to receive and manage incoming calls.

## Features

- 🔐 Agent authentication
- 📞 Real-time WebSocket call notifications
- 📊 Call statistics dashboard
- ✅ Mark calls as completed
- 🎨 Modern UI with Tailwind CSS

## Getting Started

### Install Dependencies

```bash
npm install
```

### Start Development Server

```bash
npm run dev
```

The dashboard will be available at `http://localhost:5173`

## Usage

1. **Login** with your agent ID and password
   - Example Agent IDs: `945c73`, `fba11e`, `bc13da`
   - Default password: `password123`

2. **View Calls** - Real-time updates via WebSocket
   - See assigned calls
   - View call statistics
   - Complete calls

## Build for Production

```bash
npm run build
```

## Configuration

The dashboard connects to:
- API: `http://localhost:8082`
- WebSocket: `ws://localhost:8082/ws/assigned`

Update these URLs in the components if your backend runs on different ports.