import { Component, ErrorInfo, ReactNode } from 'react'
import { AlertTriangle, RefreshCcw } from 'lucide-react'

interface Props {
  children?: ReactNode
}

interface State {
  hasError: boolean
  error: Error | null
}

export class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null,
  }

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Uncaught error:', error, errorInfo)
  }

  public render() {
    if (this.state.hasError) {
      return (
        <div className="flex h-full flex-col items-center justify-center p-6 text-center">
          <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-[var(--color-danger)]/10 text-[var(--color-danger)]">
            <AlertTriangle size={32} />
          </div>
          <h2 className="mb-2 text-xl font-bold text-[var(--color-text)]">Something went wrong</h2>
          <p className="mb-6 max-w-md text-[var(--color-text-dim)]">
            The application encountered an unexpected error. This might be due to a backend connection issue or a component failure.
          </p>
          <div className="rounded-lg bg-[var(--color-panel)] p-4 text-left font-mono text-xs text-[var(--color-danger)] border border-[var(--color-border)] mb-6 w-full max-w-lg overflow-auto">
            {this.state.error?.toString()}
          </div>
          <button
            onClick={() => window.location.reload()}
            className="flex items-center gap-2 bg-[var(--color-accent)] text-white px-4 py-2 rounded-lg hover:opacity-90 transition-opacity"
          >
            <RefreshCcw size={16} />
            Reload Application
          </button>
        </div>
      )
    }

    return this.props.children
  }
}
