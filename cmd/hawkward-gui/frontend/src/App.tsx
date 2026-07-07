import { useState } from 'react'
import { Sidebar } from './components/layout/Sidebar'
import { TopBar } from './components/layout/TopBar'
import { MainContent } from './components/layout/MainContent'

export type Page = 'dashboard' | 'sysops' | 'netops' | 'secops' | 'devops' | 'aiops' | 'network-design' | 'logs' | 'settings'

function App() {
  const [currentPage, setCurrentPage] = useState<Page>('dashboard')

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-[var(--color-bg)]">
      <Sidebar currentPage={currentPage} onNavigate={setCurrentPage} />
      <div className="flex-1 flex flex-col overflow-hidden">
        <TopBar currentPage={currentPage} />
        <MainContent currentPage={currentPage} />
      </div>
    </div>
  )
}

export default App
