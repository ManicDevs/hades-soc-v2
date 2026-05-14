import React from 'react'

export const Tabs = ({ children, defaultValue, value, onValueChange, className }: { children: React.ReactNode; defaultValue?: string; value?: string; onValueChange?: (value: string) => void; className?: string }) => {
  const [active, setActive] = React.useState(defaultValue)
  const current = value !== undefined ? value : active
  const handleChange = onValueChange || setActive

  return (
    <div data-state={current}>
      {React.Children.map(children, (child: any) => {
        if (child?.type?.displayName === 'TabsList') {
          return React.cloneElement(child as React.ReactElement, { onChange: handleChange, active })
        }
        if (child?.props?.value === current) return child
        return null
      })}
    </div>
  )
}
Tabs.displayName = 'Tabs'

export const TabsList = ({ children, active, className }: { children: React.ReactNode; active?: string; className?: string }) => (
  <div className={`flex border-b border-gray-700 mb-4 ${className || ''}`}>
    {React.Children.map(children, (child: any) => {
      const value = child?.props?.value
      const isActive = value === active
      return React.cloneElement(child as React.ReactElement, { isActive })
    })}
  </div>
)
TabsList.displayName = 'TabsList'

export const TabsTrigger = ({ children, value, isActive, onChange, className }: { children: React.ReactNode; value: string; isActive?: boolean; onChange?: (value: string) => void; className?: string }) => (
  <button
    className={`px-4 py-2 font-medium transition-colors border-b-2 -mb-px ${
      isActive ? 'border-blue-500 text-blue-400' : 'border-transparent text-gray-400 hover:text-white'
    } ${className || ''}`}
    onClick={() => onChange?.(value)}
  >
    {children}
  </button>
)
TabsTrigger.displayName = 'TabsTrigger'

export const TabsContent = ({ children, value, className }: { children: React.ReactNode; value?: string; className?: string }) => children
TabsContent.displayName = 'TabsContent'