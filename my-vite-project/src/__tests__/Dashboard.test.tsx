import { render, screen, waitFor } from './test-utils'
import Dashboard from '../pages/Dashboard'
import { mockUrlResponse } from './test-utils'

// Mock the API
jest.mock('../services/api', () => ({
  urlApi: {
    getUrls: jest.fn(),
    getUrl: jest.fn(),
    createUrl: jest.fn(),
    deleteUrl: jest.fn(),
    bulkDeleteUrls: jest.fn(),
  },
  crawlApi: {
    startCrawl: jest.fn(),
    getCrawlStatus: jest.fn(),
    bulkRerunCrawls: jest.fn(),
  },
}))

// Mock useNavigate  
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => jest.fn(),
  Link: ({ children, to }: any) => <a href={to}>{children}</a>,
}))

describe('Dashboard Component', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    
    // Mock successful API response
    const { urlApi } = require('../services/api')
    urlApi.getUrls.mockResolvedValue(mockUrlResponse)
  })

  test('renders dashboard header correctly', () => {
    render(<Dashboard />)
    
    expect(screen.getByText('Dashboard')).toBeInTheDocument()
    expect(screen.getByText('Add URL')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Search...')).toBeInTheDocument()
  })

  test('displays URLs in table after loading', async () => {
    render(<Dashboard />)
    
    // Wait for data to load
    await waitFor(() => {
      expect(screen.getByText('https://example.com')).toBeInTheDocument()
    })
    
    expect(screen.getByText('Example Website')).toBeInTheDocument()
    // Find detail links by role and href
    const detailLinks = screen.getAllByRole('link').filter(link => link.getAttribute('href')?.startsWith('/url/'))
    expect(detailLinks.length).toBeGreaterThanOrEqual(1)
  })

  test('shows View buttons for URLs', async () => {
    render(<Dashboard />)
    
    // Wait for data to load
    await waitFor(() => {
      const links = screen.getAllByRole('link').filter(link => link.getAttribute('href')?.startsWith('/url/'))
      expect(links.length).toBeGreaterThanOrEqual(1)
    })
    
    const detailLinks = screen.getAllByRole('link').filter(link => link.getAttribute('href')?.startsWith('/url/'))
    expect(detailLinks[0]).toHaveAttribute('href', '/url/1')
    expect(detailLinks[1]).toHaveAttribute('href', '/url/2')
  })

  test('displays loading state initially', () => {
    render(<Dashboard />)
    // Should show skeleton loading cards
    expect(document.querySelectorAll('.animate-pulse').length).toBeGreaterThan(0)
  })
}) 