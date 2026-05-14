export const Card = ({ children, className = '' }: { children: React.ReactNode; className?: string }) => (
  <div className={`bg-gray-800 rounded-lg border border-gray-700 ${className}`}>
    {children}
  </div>
)
export const CardHeader = ({ children, className }: { children: React.ReactNode; className?: string }) => (
  <div className="px-4 py-3 border-b border-gray-700">{children}</div>
)
export const CardTitle = ({ children, className }: { children: React.ReactNode; className?: string }) => (
  <h3 className={`text-lg font-semibold text-white ${className || ''}`}>{children}</h3>
)
export const CardContent = ({ children, className }: { children: React.ReactNode; className?: string }) => (
  <div className={`p-4 ${className || ''}`}>{children}</div>
)