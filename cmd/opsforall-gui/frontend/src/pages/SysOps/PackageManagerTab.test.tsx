import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { PackageManagerTab } from './PackageManagerTab'
import { toast } from 'sonner'

// ── Mocks ────────────────────────────────────────────────────────────────────

const mockPackages = [
  { name: 'curl', version: '7.76.1' },
  { name: 'git', version: '2.35.1' },
  { name: 'vim', version: '8.2' },
  { name: 'node.js', version: '18.12.0' },
  { name: 'python', version: '3.12.0' },
]

type MockData = { name: string; version: string }[]

let mockManagers: { name: string; found: boolean; packages: MockData }[]
vi.mock('@tanstack/react-query', () => ({
  useQuery: () => {
    return { data: mockManagers, isLoading: false, isSuccess: true }
  },
  useQueryClient: () => ({ invalidateQueries: vi.fn() }),
}))

vi.mock('@/hooks/useBackend', () => ({
  useBackend: () => ({
    call: vi.fn().mockResolvedValue([]),
  }),
}))

vi.mock('@tanstack/react-virtual', () => ({
  useVirtualizer: ({ count, estimateSize }: any) => ({
    getVirtualItems: () =>
      Array.from({ length: count }, (_, i) => ({
        key: i,
        index: i,
        start: estimateSize() * i,
        size: estimateSize(),
        measureElement: vi.fn(),
      })),
    getTotalSize: () => count * estimateSize(),
    measureElement: vi.fn(),
  }),
}))

vi.mock('@/stores/useSettingsStore', () => ({
  useSettingsStore: () => ({ refreshInterval: undefined }),
}))

vi.mock('sonner', () => ({
  toast: { success: vi.fn() },
}))

// ── Helpers ──────────────────────────────────────────────────────────────────

beforeEach(() => {
  mockManagers = []
  vi.clearAllMocks()
})

// ── Tests ────────────────────────────────────────────────────────────────────

describe('PackageManagerTab', () => {
  it('shows "None detected" when no manager is found', () => {
    mockManagers = [
      { name: 'winget', found: false, packages: [] },
      { name: 'choco', found: false, packages: [] },
    ]
    render(<PackageManagerTab />)
    expect(screen.getByText('None detected')).toBeInTheDocument()
    expect(screen.getByText(/install winget/i)).toBeInTheDocument()
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
  })

  it('shows "None detected" when managers array is empty', () => {
    mockManagers = []
    render(<PackageManagerTab />)
    expect(screen.getByText('None detected')).toBeInTheDocument()
    expect(screen.getByText(/install winget/i)).toBeInTheDocument()
  })

  it('renders first found manager badge and table', () => {
    mockManagers = [
      { name: 'winget', found: false, packages: [] },
      { name: 'windows-installed', found: true, packages: mockPackages },
    ]
    render(<PackageManagerTab />)
    expect(screen.getByText('Windows Apps (Registry)')).toBeInTheDocument()
    expect(screen.getByRole('table')).toBeInTheDocument()
  })

  it('displays all packages with row numbers', () => {
    mockManagers = [
      { name: 'winget', found: true, packages: mockPackages },
    ]
    render(<PackageManagerTab />)
    const rows = screen.getAllByTestId('package-row')
    expect(rows).toHaveLength(5)
    expect(screen.getAllByTestId('row-number').map(el => el.textContent)).toEqual(['1', '2', '3', '4', '5'])
  })

  it('provides a positioned virtualized table body with total height', () => {
    mockManagers = [{ name: 'winget', found: true, packages: mockPackages }]
    render(<PackageManagerTab />)

    const body = screen.getByRole('table').querySelector('tbody') as HTMLTableSectionElement
    expect(body).toHaveStyle({ position: 'relative', height: '205px', width: '100%' })
  })

  it('shows search input and filters packages by name', async () => {
    mockManagers = [{ name: 'winget', found: true, packages: mockPackages }]
    render(<PackageManagerTab />)
    const input = screen.getByPlaceholderText(/search by name/i)
    expect(input).toBeInTheDocument()

    fireEvent.change(input, { target: { value: 'git' } })
    await waitFor(() => {
      const rows = screen.getAllByTestId('package-row')
      expect(rows).toHaveLength(1)
      expect(screen.getByText('git')).toBeInTheDocument()
    })
  })

  it('filters by version string', async () => {
    mockManagers = [{ name: 'winget', found: true, packages: mockPackages }]
    render(<PackageManagerTab />)
    const input = screen.getByPlaceholderText(/search by name/i)
    fireEvent.change(input, { target: { value: '7.76' } })
    await waitFor(() => {
      const rows = screen.getAllByTestId('package-row')
      expect(rows).toHaveLength(1)
      expect(screen.getByText('curl')).toBeInTheDocument()
    })
  })

  it('shows "does not match your search" when filter yields nothing', async () => {
    mockManagers = [{ name: 'winget', found: true, packages: mockPackages }]
    render(<PackageManagerTab />)
    const input = screen.getByPlaceholderText(/search by name/i)
    fireEvent.change(input, { target: { value: 'zzz_no_match' } })
    await waitFor(() => {
      expect(screen.getByText(/No packages match/i)).toBeInTheDocument()
    })
  })

  it('clear button resets search and shows all packages', async () => {
    mockManagers = [{ name: 'winget', found: true, packages: mockPackages }]
    render(<PackageManagerTab />)
    const input = screen.getByPlaceholderText(/search by name/i)
    fireEvent.change(input, { target: { value: 'git' } })
    await waitFor(() => {
      expect(screen.getAllByTestId('package-row')).toHaveLength(1)
    })
    // Clear button should appear
    const clearBtn = screen.getByTitle('Clear')
    expect(clearBtn).toBeInTheDocument()
    fireEvent.click(clearBtn)
    await waitFor(() => {
      expect(screen.getAllByTestId('package-row')).toHaveLength(5)
      expect(input).toHaveValue('')
    })
  })

  it('sorts by name ascending by default', () => {
    mockManagers = [{ name: 'winget', found: true, packages: mockPackages }]
    render(<PackageManagerTab />)
    const names = screen.getAllByTestId('package-name').map(el => el.textContent)
    // localeCompare with numeric: true sorts: curl, git, node.js, python, vim
    expect(names).toEqual(['curl', 'git', 'node.js', 'python', 'vim'])
  })

  it('toggles sort direction when clicking name column', async () => {
    mockManagers = [{ name: 'winget', found: true, packages: mockPackages }]
    render(<PackageManagerTab />)
    // Click name sort button to toggle to desc
    const nameBtn = screen.getByText('Package').closest('button')!
    fireEvent.click(nameBtn)
    await waitFor(() => {
      const names = screen.getAllByTestId('package-name').map(el => el.textContent)
      expect(names).toEqual(['vim', 'python', 'node.js', 'git', 'curl'])
    })
    // Click again to go back to asc
    fireEvent.click(nameBtn)
    await waitFor(() => {
      const names = screen.getAllByTestId('package-name').map(el => el.textContent)
      expect(names).toEqual(['curl', 'git', 'node.js', 'python', 'vim'])
    })
  })

  it('sorts by version when clicking version column', async () => {
    mockManagers = [{ name: 'winget', found: true, packages: mockPackages }]
    render(<PackageManagerTab />)
    const versionBtn = screen.getByText('Version').closest('button')!
    fireEvent.click(versionBtn)
    await waitFor(() => {
      const names = screen.getAllByTestId('package-name').map(el => el.textContent)
      // Versions: 2.35.1 (git), 7.76.1 (curl), 8.2 (vim), 18.12.0 (node.js), 3.12.0 (python)
      // Sorted by version asc with numeric: true: 2.35.1, 3.12.0, 7.76.1, 8.2, 18.12.0
      expect(names).toEqual(['git', 'python', 'curl', 'vim', 'node.js'])
    })
  })

  it('copy button copies name to clipboard and shows toast', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined)
    Object.assign(navigator, { clipboard: { writeText } })

    mockManagers = [{ name: 'winget', found: true, packages: [mockPackages[0]] }]
    render(<PackageManagerTab />)
    const copyBtn = screen.getByTestId('copy-button')
    fireEvent.click(copyBtn)
    await waitFor(() => {
      expect(writeText).toHaveBeenCalledWith('curl')
      expect(toast.success).toHaveBeenCalledWith('Copied "curl"')
    })
  })

  it('version badge renders for packages with version', () => {
    mockManagers = [{ name: 'winget', found: true, packages: mockPackages }]
    render(<PackageManagerTab />)
    const badges = screen.getAllByTestId('version-badge')
    expect(badges).toHaveLength(5)
    expect(badges[0]).toHaveTextContent('7.76.1')
  })

  it('shows dash for packages without version', () => {
    mockManagers = [{
      name: 'winget', found: true,
      packages: [{ name: 'test-pkg', version: '' }],
    }]
    render(<PackageManagerTab />)
    expect(screen.getByText('—')).toBeInTheDocument()
  })

  it('shows filtered count when search is active', async () => {
    mockManagers = [{ name: 'winget', found: true, packages: mockPackages }]
    render(<PackageManagerTab />)
    const input = screen.getByPlaceholderText(/search by name/i)
    fireEvent.change(input, { target: { value: 'git' } })
    await waitFor(() => {
      expect(screen.getByText('filtered')).toBeInTheDocument()
    })
  })

  it('shows footer stats with sort info', () => {
    mockManagers = [{ name: 'winget', found: true, packages: mockPackages }]
    render(<PackageManagerTab />)
    expect(screen.getByText(/sorted by/i)).toBeInTheDocument()
    expect(screen.getByText(/A→Z/)).toBeInTheDocument()
    // The footer shows "Showing 5 of 5 packages" — find the whole span
    const footer = screen.getByText((content, element) =>
      content.includes('Showing') && element?.tagName === 'SPAN' && content.includes('packages')
    )
    expect(footer).toBeInTheDocument()
  })

  it('shows total package count in header', () => {
    mockManagers = [{ name: 'winget', found: true, packages: mockPackages }]
    render(<PackageManagerTab />)
    // There are multiple "5" elements; the header one is in the manager badge area
    const counts = screen.getAllByText('5')
    expect(counts.length).toBeGreaterThanOrEqual(1)
    expect(screen.getByText('packages')).toBeInTheDocument()
  })
})
