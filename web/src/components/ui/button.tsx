export const Button = ({ children, className = '', variant = 'default', ...props }: { children: React.ReactNode; className?: string; variant?: 'default' | 'secondary'; [key: string]: any }) => {
  const variants = {
    default: 'bg-blue-600 hover:bg-blue-700 text-white',
    secondary: 'bg-gray-700 hover:bg-gray-600 text-white',
  }
  return (
    <button
      className={`px-4 py-2 rounded font-medium transition-colors ${variants[variant]} ${className}`}
      {...props}
    >
      {children}
    </button>
  )
}