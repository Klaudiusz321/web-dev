import React from 'react'
import Sidebar from './Sidebar'
import CommandPalette, { useCommandPalette } from './CommandPalette'

interface AppLayoutProps {
  children: React.ReactNode
}

export default function AppLayout({ children }: AppLayoutProps) {
  const commandPalette = useCommandPalette()

  return (
    <div className="flex h-screen bg-bg-primary overflow-hidden">
      {/* Sidebar */}
      <Sidebar />
      
      {/* Main Content */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Main Area */}
        <main className="flex-1 overflow-y-auto bg-bg-primary">
          <div className="min-h-full">
            {children}
          </div>
        </main>
      </div>

      {/* Command Palette */}
      <CommandPalette 
        isOpen={commandPalette.isOpen} 
        onClose={commandPalette.close} 
      />
    </div>
  )
} 