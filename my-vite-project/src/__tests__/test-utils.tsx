import React from 'react'
import type { ReactElement } from 'react'
import { render } from '@testing-library/react'
import type { RenderOptions } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { Toaster } from 'react-hot-toast'

// Create a custom render function that includes providers
const AllTheProviders = ({ children }: { children: React.ReactNode }) => {
  // Create a new QueryClient for each test to ensure isolation
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        staleTime: 0,
        gcTime: 0,
      },
      mutations: {
        retry: false,
      },
    },
  })

  return (
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        {children}
        <Toaster position="top-right" />
      </MemoryRouter>
    </QueryClientProvider>
  )
}

const customRender = (
  ui: ReactElement,
  options?: Omit<RenderOptions, 'wrapper'>
) => render(ui, { wrapper: AllTheProviders, ...options })

// Custom render with specific router state
export const renderWithRouter = (
  ui: ReactElement,
  { initialEntries = ['/'], ...options }: { initialEntries?: string[] } & Omit<RenderOptions, 'wrapper'> = {}
) => {
  const Wrapper = ({ children }: { children: React.ReactNode }) => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
          staleTime: 0,
          gcTime: 0,
        },
        mutations: {
          retry: false,
        },
      },
    })

    return (
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={initialEntries}>
          {children}
          <Toaster position="top-right" />
        </MemoryRouter>
      </QueryClientProvider>
    )
  }

  return render(ui, { wrapper: Wrapper, ...options })
}

// Mock API responses
export const mockUrlResponse = {
  data: [
    {
      id: 1,
      url: 'https://example.com',
      title: 'Example Website',
      html_version: 'HTML5',
      status: 'completed' as const,
      has_login_form: false,
      created_at: '2025-01-01T12:00:00Z',
      updated_at: '2025-01-01T12:00:00Z',
    },
    {
      id: 2,
      url: 'https://test.com',
      title: 'Test Site',
      html_version: 'HTML5',
      status: 'pending' as const,
      has_login_form: true,
      created_at: '2025-01-01T11:00:00Z',
      updated_at: '2025-01-01T11:00:00Z',
    },
  ],
  pagination: {
    total: 2,
    limit: 10,
    offset: 0,
  },
}

export const mockSingleUrlResponse = {
  data: {
    id: 1,
    url: 'https://example.com',
    title: 'Example Website',
    html_version: 'HTML5',
    status: 'completed' as const,
    has_login_form: false,
    created_at: '2025-01-01T12:00:00Z',
    updated_at: '2025-01-01T12:00:00Z',
  },
}

export const mockCrawlStatusResponse = {
  data: {
    id: 1,
    url: 'https://example.com',
    status: 'completed',
    internal_links: 25,
    external_links: 8,
    broken_links: 2,
    heading_counts: {
      h1: 1,
      h2: 5,
      h3: 12,
      h4: 3,
      h5: 0,
      h6: 0,
    },
    started_at: '2025-01-01T12:00:00Z',
    completed_at: '2025-01-01T12:05:00Z',
  },
}

export const mockLinksResponse = {
  data: [
    {
      id: 1,
      url_id: 1,
      crawl_id: 1,
      link_url: 'https://example.com/about',
      link_text: 'About Us',
      link_type: 'internal' as const,
      status_code: 200,
      is_accessible: true,
      created_at: '2025-01-01T12:00:00Z',
    },
    {
      id: 2,
      url_id: 1,
      crawl_id: 1,
      link_url: 'https://broken-link.com',
      link_text: 'Broken Link',
      link_type: 'external' as const,
      status_code: 404,
      is_accessible: false,
      created_at: '2025-01-01T12:00:00Z',
    },
  ],
  pagination: {
    total: 2,
    limit: 10,
    offset: 0,
  },
}

// Re-export everything from React Testing Library
export * from '@testing-library/react'

// Override render method
export { customRender as render } 