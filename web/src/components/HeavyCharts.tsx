import React, { Suspense, useState, useEffect } from 'react'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  AreaChart,
  Area,
} from 'recharts'
import { SuspenseFallback } from './SuspenseFallback'

interface ChartWrapperProps {
  children: React.ReactNode
  height?: number
  title?: string
}

const ChartWrapper: React.FC<ChartWrapperProps> = ({ children, height = 300, title }) => (
  <div className="hades-card p-4">
    {title && <h3 className="text-white font-medium mb-4">{title}</h3>}
    <Suspense fallback={<SuspenseFallback message="Loading chart..." size="sm" />}>
      <div style={{ height }}>
        {children}
      </div>
    </Suspense>
  </div>
)

export const SecurityOverviewChart: React.FC<{ data?: unknown[] }> = ({ data = [] }) => {
  const [chartData, setChartData] = useState(data)

  useEffect(() => {
    if (data.length === 0) {
      const generatedData = Array.from({ length: 24 }, (_, i) => ({
        hour: `${i}:00`,
        threats: Math.floor(Math.random() * 20),
        blocked: Math.floor(Math.random() * 15),
        monitoring: Math.floor(Math.random() * 10)
      }))
      setChartData(generatedData)
    }
  }, [data])

  return (
    <ChartWrapper title="Security Overview (24h)">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={chartData}>
          <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
          <XAxis dataKey="hour" stroke="#64748b" fontSize={12} />
          <YAxis stroke="#64748b" fontSize={12} />
          <Tooltip
            contentStyle={{
              backgroundColor: '#1e293b',
              border: '1px solid #334155',
              borderRadius: '8px'
            }}
            labelStyle={{ color: '#fff' }}
          />
          <Legend />
          <Line type="monotone" dataKey="threats" stroke="#ef4444" strokeWidth={2} />
          <Line type="monotone" dataKey="blocked" stroke="#22c55e" strokeWidth={2} />
          <Line type="monotone" dataKey="monitoring" stroke="#f59e0b" strokeWidth={2} />
        </LineChart>
      </ResponsiveContainer>
    </ChartWrapper>
  )
}

export const ThreatDistributionChart: React.FC<{ data?: unknown[] }> = ({ data = [] }) => {
  const defaultColors = ['#ef4444', '#f59e0b', '#22c55e', '#3b82f6', '#8b5cf6']

  const defaultData = [
    { name: 'Critical', value: 12 },
    { name: 'High', value: 45 },
    { name: 'Medium', value: 78 },
    { name: 'Low', value: 156 },
    { name: 'Info', value: 234 }
  ]

  const chartData = data.length > 0 ? data : defaultData

  return (
    <ChartWrapper title="Threat Distribution">
      <ResponsiveContainer width="100%" height="100%">
        <PieChart>
          <Pie
            data={chartData}
            cx="50%"
            cy="50%"
            innerRadius={40}
            outerRadius={80}
            paddingAngle={2}
            dataKey="value"
            label={({ name, percent }: { name?: string; percent?: number }) => `${name ?? ''} ${((percent ?? 0) * 100).toFixed(0)}%`}
          >
            {chartData.map((_: unknown, index: number) => (
              <Cell key={`cell-${index}`} fill={defaultColors[index % defaultColors.length]} />
            ))}
          </Pie>
          <Tooltip
            contentStyle={{
              backgroundColor: '#1e293b',
              border: '1px solid #334155',
              borderRadius: '8px'
            }}
          />
        </PieChart>
      </ResponsiveContainer>
    </ChartWrapper>
  )
}

export const BlockedAttacksChart: React.FC<{ data?: unknown[] }> = ({ data = [] }) => {
  const [chartData, setChartData] = useState(data)

  useEffect(() => {
    if (data.length === 0) {
      const generatedData = Array.from({ length: 7 }, (_, i) => {
        const day = new Date()
        day.setDate(day.getDate() - (6 - i))
        return {
          day: day.toLocaleDateString('en-US', { weekday: 'short' }),
          blocked: Math.floor(Math.random() * 500) + 200,
          allowed: Math.floor(Math.random() * 100) + 50
        }
      })
      setChartData(generatedData)
    }
  }, [data])

  return (
    <ChartWrapper title="Blocked vs Allowed (7 days)">
      <ResponsiveContainer width="100%" height="100%">
        <BarChart data={chartData}>
          <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
          <XAxis dataKey="day" stroke="#64748b" fontSize={12} />
          <YAxis stroke="#64748b" fontSize={12} />
          <Tooltip
            contentStyle={{
              backgroundColor: '#1e293b',
              border: '1px solid #334155',
              borderRadius: '8px'
            }}
          />
          <Legend />
          <Bar dataKey="blocked" fill="#22c55e" radius={[4, 4, 0, 0]} />
          <Bar dataKey="allowed" fill="#64748b" radius={[4, 4, 0, 0]} />
        </BarChart>
      </ResponsiveContainer>
    </ChartWrapper>
  )
}

export const NetworkTrafficChart: React.FC<{ data?: unknown[] }> = ({ data = [] }) => {
  const [chartData, setChartData] = useState(data)

  useEffect(() => {
    if (data.length === 0) {
      const generatedData = Array.from({ length: 24 }, (_, i) => ({
        hour: `${i}:00`,
        inbound: Math.floor(Math.random() * 1000) + 200,
        outbound: Math.floor(Math.random() * 800) + 100
      }))
      setChartData(generatedData)
    }
  }, [data])

  return (
    <ChartWrapper title="Network Traffic (24h)">
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={chartData}>
          <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
          <XAxis dataKey="hour" stroke="#64748b" fontSize={12} />
          <YAxis stroke="#64748b" fontSize={12} />
          <Tooltip
            contentStyle={{
              backgroundColor: '#1e293b',
              border: '1px solid #334155',
              borderRadius: '8px'
            }}
          />
          <Area
            type="monotone"
            dataKey="inbound"
            stroke="#3b82f6"
            fill="#3b82f6"
            fillOpacity={0.3}
          />
          <Area
            type="monotone"
            dataKey="outbound"
            stroke="#22c55e"
            fill="#22c55e"
            fillOpacity={0.3}
          />
        </AreaChart>
      </ResponsiveContainer>
    </ChartWrapper>
  )
}

interface HeavyChartsProps {
  type?: 'security' | 'threats' | 'network' | 'all'
  className?: string
}

export const HeavyCharts: React.FC<HeavyChartsProps> = ({ type = 'all', className = '' }) => {
  return (
    <div className={`grid grid-cols-1 lg:grid-cols-2 gap-6 ${className}`}>
      {type === 'security' || type === 'all' ? (
        <>
          <SecurityOverviewChart />
          <ThreatDistributionChart />
        </>
      ) : null}
      {type === 'threats' || type === 'all' ? (
        <>
          <BlockedAttacksChart />
        </>
      ) : null}
      {type === 'network' || type === 'all' ? (
        <>
          <NetworkTrafficChart />
        </>
      ) : null}
    </div>
  )
}

export default HeavyCharts