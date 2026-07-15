import React from 'react'
import ReactDOM from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { Toaster } from 'sonner'
import App from './App'
import './styles/globals.css'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchInterval: false,       // we control polling per-page
      retry: 1,
      staleTime: 5000,              // 5s — reduces re-renders from identical refetches
      refetchOnWindowFocus: false,
    },
  },
})

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <App />
      <Toaster
        position="bottom-right"
        richColors
        closeButton
        expand
        toastOptions={{
          duration: 5000,
        }}
      />
    </QueryClientProvider>
  </React.StrictMode>
)
