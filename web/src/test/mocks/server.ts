import { setupServer } from 'msw/node'
import { http } from 'msw'

// Mock API handlers
const handlers = [
  // Auth endpoints
  http.post('/api/auth/login', () => {
    return new Response(JSON.stringify({
      success: true,
      data: {
        user: {
          id: '1',
          username: 'testuser',
          email: 'test@example.com',
          role: 'admin',
          permissions: ['read', 'write', 'admin']
        },
        token: 'mock-jwt-token'
      }
    }), {
      headers: { 'Content-Type': 'application/json' }
    })
  }),

  http.post('/api/auth/logout', () => {
    return new Response(JSON.stringify({ success: true }), {
      headers: { 'Content-Type': 'application/json' }
    })
  }),

  // Dashboard endpoints
  http.get('/api/dashboard/stats', () => {
    return new Response(JSON.stringify({
      success: true,
      data: {
        totalThreats: 42,
        activeScans: 3,
        blockedAttacks: 128,
        systemHealth: 98
      }
    }), {
      headers: { 'Content-Type': 'application/json' }
    })
  }),

  // Threat endpoints
  http.get('/api/threats', () => {
    return new Response(JSON.stringify({
      success: true,
      data: {
        threats: [
          {
            id: '1',
            type: 'malware',
            severity: 'high',
            status: 'active',
            description: 'Suspicious malware detected',
            timestamp: '2024-01-15T10:30:00Z'
          }
        ],
        total: 1
      }
    }), {
      headers: { 'Content-Type': 'application/json' }
    })
  }),

  // Health check
  http.get('/api/health', () => {
    return new Response(JSON.stringify({
      status: 'healthy',
      timestamp: new Date().toISOString()
    }), {
      headers: { 'Content-Type': 'application/json' }
    })
  })
]

export const server = setupServer(...handlers)
