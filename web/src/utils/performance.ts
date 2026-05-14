import React from 'react'

// Performance monitoring utilities for Hades-V2 frontend

interface PerformanceMetrics {
  renderTime: number
  componentLoadTime: number
  apiResponseTime: number
  memoryUsage: number
  errorCount: number
}

interface PerformanceEntry {
  name: string
  startTime: number
  duration: number
  type: 'render' | 'api' | 'component' | 'navigation'
}

class PerformanceMonitor {
  private metrics: PerformanceMetrics
  private entries: PerformanceEntry[] = []
  private observers: PerformanceObserver[] = []

  constructor() {
    this.metrics = {
      renderTime: 0,
      componentLoadTime: 0,
      apiResponseTime: 0,
      memoryUsage: 0,
      errorCount: 0,
    }
    this.initializeObservers()
  }

  private initializeObservers() {
    // Observer for paint timing
    if ('PerformanceObserver' in window) {
      const paintObserver = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (entry.name === 'first-contentful-paint') {
            this.recordMetric('renderTime', entry.startTime)
          }
        }
      })
      paintObserver.observe({ entryTypes: ['paint'] })
      this.observers.push(paintObserver)

      // Observer for navigation timing
      const navigationObserver = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (entry.entryType === 'navigation') {
            const navEntry = entry as PerformanceNavigationTiming
            this.recordMetric('renderTime', navEntry.loadEventEnd - navEntry.loadEventStart)
          }
        }
      })
      navigationObserver.observe({ entryTypes: ['navigation'] })
      this.observers.push(navigationObserver)

      // Observer for resource timing
      const resourceObserver = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (entry.entryType === 'resource') {
            const resourceEntry = entry as PerformanceResourceTiming
            if (resourceEntry.initiatorType === 'fetch' || resourceEntry.initiatorType === 'xmlhttprequest') {
              this.recordMetric('apiResponseTime', resourceEntry.responseEnd - resourceEntry.requestStart)
            }
          }
        }
      })
      resourceObserver.observe({ entryTypes: ['resource'] })
      this.observers.push(resourceObserver)
    }

    // Monitor memory usage
    if ('memory' in performance) {
      setInterval(() => {
        this.recordMetric('memoryUsage', (performance as any).memory.usedJSHeapSize)
      }, 5000)
    }

    // Monitor errors
    window.addEventListener('error', () => {
      this.metrics.errorCount++
    })
  }

  recordMetric(type: keyof PerformanceMetrics, value: number) {
    if (type in this.metrics) {
      this.metrics[type] = value
    }
  }

  startTimer(name: string, type: PerformanceEntry['type'] = 'component'): () => void {
    const startTime = performance.now()
    
    return () => {
      const duration = performance.now() - startTime
      this.entries.push({
        name,
        startTime,
        duration,
        type,
      })

      // Update aggregate metrics
      if (type === 'render') {
        this.metrics.renderTime = Math.max(this.metrics.renderTime, duration)
      } else if (type === 'api') {
        this.metrics.apiResponseTime = Math.max(this.metrics.apiResponseTime, duration)
      } else if (type === 'component') {
        this.metrics.componentLoadTime = Math.max(this.metrics.componentLoadTime, duration)
      }
    }
  }

  getMetrics(): PerformanceMetrics {
    return { ...this.metrics }
  }

  getEntries(): PerformanceEntry[] {
    return [...this.entries]
  }

  getAverageRenderTime(): number {
    const renderEntries = this.entries.filter(e => e.type === 'render')
    if (renderEntries.length === 0) return 0
    return renderEntries.reduce((sum, entry) => sum + entry.duration, 0) / renderEntries.length
  }

  getAverageApiResponseTime(): number {
    const apiEntries = this.entries.filter(e => e.type === 'api')
    if (apiEntries.length === 0) return 0
    return apiEntries.reduce((sum, entry) => sum + entry.duration, 0) / apiEntries.length
  }

  getSlowComponents(threshold: number = 100): PerformanceEntry[] {
    return this.entries
      .filter(e => e.type === 'component' && e.duration > threshold)
      .sort((a, b) => b.duration - a.duration)
  }

  getSlowAPIs(threshold: number = 1000): PerformanceEntry[] {
    return this.entries
      .filter(e => e.type === 'api' && e.duration > threshold)
      .sort((a, b) => b.duration - a.duration)
  }

  generateReport(): string {
    return `
Performance Report for Hades-V2 Frontend
========================================
Average Render Time: ${this.getAverageRenderTime().toFixed(2)}ms
Average API Response Time: ${this.getAverageApiResponseTime().toFixed(2)}ms
Memory Usage: ${(this.metrics.memoryUsage / 1024 / 1024).toFixed(2)}MB
Error Count: ${this.metrics.errorCount}

Slow Components (>100ms):
${this.getSlowComponents().map(c => `- ${c.name}: ${c.duration.toFixed(2)}ms`).join('\n') || 'None'}

Slow APIs (>1000ms):
${this.getSlowAPIs().map(a => `- ${a.name}: ${a.duration.toFixed(2)}ms`).join('\n') || 'None'}
    `.trim()
  }

  cleanup() {
    this.observers.forEach(observer => observer.disconnect())
    this.observers = []
    this.entries = []
  }
}

// Singleton instance
export const performanceMonitor = new PerformanceMonitor()

// React Hook for performance monitoring
export function usePerformanceMonitor(componentName: string) {
  const startTimer = () => performanceMonitor.startTimer(componentName, 'component')
  
  return {
    startTimer,
    getMetrics: performanceMonitor.getMetrics.bind(performanceMonitor),
    generateReport: performanceMonitor.generateReport.bind(performanceMonitor),
  }
}

// API wrapper with performance monitoring
export function withPerformanceMonitoring<T extends (...args: any[]) => Promise<any>>(
  fn: T,
  name: string
): T {
  return (async (...args: Parameters<T>) => {
    const endTimer = performanceMonitor.startTimer(name, 'api')
    try {
      const result = await fn(...args)
      endTimer()
      return result
    } catch (error) {
      endTimer()
      throw error
    }
  }) as T
}

// Component wrapper for performance monitoring
export function withComponentPerformanceMonitoring<P extends object>(
  Component: React.ComponentType<P>,
  componentName: string
): React.ComponentType<P> {
  return function MonitoredComponent(props: P) {
    React.useEffect(() => {
      const endTimer = performanceMonitor.startTimer(componentName, 'component')
      
      return () => {
        endTimer()
      }
    }, [])
    
    return React.createElement(Component, props)
  }
}

export default performanceMonitor
