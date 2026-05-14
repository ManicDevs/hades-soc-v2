import '@testing-library/jest-dom'
import { vi, beforeEach } from 'vitest'

// Mock window.location
Object.defineProperty(window, 'location', {
  value: {
    hostname: 'localhost',
    protocol: 'http:',
    host: 'localhost:3000',
    href: 'http://localhost:3000',
    port: '3000',
    pathname: '/',
    search: '',
    hash: '',
    assign: vi.fn(),
    replace: vi.fn(),
    reload: vi.fn()
  },
  writable: true
})

// Mock fetch
const mockFetch = vi.fn()
global.fetch = mockFetch

// Mock WebSocket
class MockWebSocket {
  static CONNECTING = 0
  static OPEN = 1
  static CLOSING = 2
  static CLOSED = 3

  readyState = MockWebSocket.CONNECTING
  onopen: ((event: Event) => void) | null = null
  onclose: ((event: CloseEvent) => void) | null = null
  onerror: ((event: Event) => void) | null = null
  onmessage: ((event: MessageEvent) => void) | null = null

  constructor(public url: string) {
    setTimeout(() => {
      this.readyState = MockWebSocket.OPEN
      this.onopen?.(new Event('open'))
    }, 10)
  }

  send(_data: string): void {
    // Mock implementation
  }

  close(_code?: number, _reason?: string): void {
    this.readyState = MockWebSocket.CLOSED
    this.onclose?.(new CloseEvent('close'))
  }
}

global.WebSocket = MockWebSocket as unknown as typeof WebSocket

// Mock window.hadesToken for auth
Object.defineProperty(window, 'hadesToken', {
  value: null,
  writable: true,
  configurable: true
})

Object.defineProperty(window, 'hadesUser', {
  value: null,
  writable: true,
  configurable: true
})

Object.defineProperty(window, 'hadesRole', {
  value: null,
  writable: true,
  configurable: true
})

Object.defineProperty(window, 'hadesEnvironment', {
  value: null,
  writable: true,
  configurable: true
})

// Mock window.open
window.open = vi.fn()

// Mock ResizeObserver
class ResizeObserverMock {
  observe = vi.fn()
  unobserve = vi.fn()
  disconnect = vi.fn()
}
global.ResizeObserver = ResizeObserverMock

// Mock IntersectionObserver
class IntersectionObserverMock {
  observe = vi.fn()
  unobserve = vi.fn()
  disconnect = vi.fn()
  takeRecords = vi.fn()
  constructor(_callback: IntersectionObserverCallback, _options?: IntersectionObserverInit) {
    // Mock constructor
  }
}
global.IntersectionObserver = IntersectionObserverMock as any

// Reset mocks before each test
beforeEach(() => {
  vi.clearAllMocks()
  mockFetch.mockReset()
  window.hadesToken = null
  window.hadesUser = null
  window.hadesRole = null
  window.hadesEnvironment = null
})
