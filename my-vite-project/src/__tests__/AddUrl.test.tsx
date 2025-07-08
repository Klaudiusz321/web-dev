import { render, screen, waitFor } from './test-utils'
import userEvent from '@testing-library/user-event'
import AddUrl from '../pages/AddUrl'
import { urlApi } from '../services/api'

// Mock the API
jest.mock('../services/api', () => ({
  urlApi: {
    createUrl: jest.fn(),
  },
}))

// Mock useNavigate
const mockNavigate = jest.fn()
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockNavigate,
  Link: ({ children, to }: any) => <a href={to}>{children}</a>,
}))

const mockCreateUrl = urlApi.createUrl as jest.MockedFunction<typeof urlApi.createUrl>

describe('AddUrl Component', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  test('renders add URL form correctly', () => {
    render(<AddUrl />)
    
    expect(screen.getByText('Add New URL')).toBeInTheDocument()
    expect(screen.getByText('Start crawling and analyzing any website with our AI-powered web crawler')).toBeInTheDocument()
    expect(screen.getByLabelText(/website url/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /start web analysis/i })).toBeInTheDocument()
    expect(screen.getByText('Back to Dashboard')).toBeInTheDocument()
  })

  test('validates URL input correctly', async () => {
    const user = userEvent.setup()
    render(<AddUrl />)
    
    const input = screen.getByLabelText(/website url/i)
    const submitButton = screen.getByRole('button', { name: /start web analysis/i })
    
    // Test empty submission (button should be disabled)
    expect(submitButton).toBeDisabled()
    
    // Test invalid URL
    await user.type(input, 'invalid-url')
    await user.click(submitButton)
    expect(screen.getByText('Please enter a valid URL starting with http:// or https://')).toBeInTheDocument()
    
    // Clear input
    await user.clear(input)
    
    // Test URL without protocol
    await user.type(input, 'example.com')
    await user.click(submitButton)
    expect(screen.getByText('Please enter a valid URL starting with http:// or https://')).toBeInTheDocument()
  })

  test('successfully submits valid URL and redirects', async () => {
    const user = userEvent.setup()
    
    // Mock successful API response
    mockCreateUrl.mockResolvedValueOnce({
      data: {
        id: 1,
        url: 'https://example.com',
        title: '',
        html_version: '',
        status: 'pending',
        has_login_form: false,
        created_at: '2025-01-01T12:00:00Z',
        updated_at: '2025-01-01T12:00:00Z',
      },
    })
    
    render(<AddUrl />)
    
    const input = screen.getByLabelText(/website url/i)
    const submitButton = screen.getByRole('button', { name: /start web analysis/i })
    
    // Enter valid URL
    await user.type(input, 'https://example.com')
    await user.click(submitButton)
    
    // Wait for API call
    await waitFor(() => {
      expect(mockCreateUrl).toHaveBeenCalledWith('https://example.com')
    })
    
    // Should redirect to dashboard
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/')
    })
  })

  test('handles API error gracefully', async () => {
    const user = userEvent.setup()
    
    // Mock API error
    mockCreateUrl.mockRejectedValueOnce(new Error('Failed to add URL'))
    
    render(<AddUrl />)
    
    const input = screen.getByLabelText(/website url/i)
    const submitButton = screen.getByRole('button', { name: /start web analysis/i })
    
    // Enter valid URL
    await user.type(input, 'https://example.com')
    await user.click(submitButton)
    
    // Wait for API call to fail
    await waitFor(() => {
      expect(mockCreateUrl).toHaveBeenCalledWith('https://example.com')
    })
    
    // Should not redirect on error
    expect(mockNavigate).not.toHaveBeenCalled()
  })

  test('shows loading state during submission', async () => {
    const user = userEvent.setup()
    
    // Mock delayed API response
    mockCreateUrl.mockImplementationOnce(
      () => new Promise(resolve => setTimeout(() => resolve({
        data: {
          id: 1,
          url: 'https://example.com',
          title: '',
          html_version: '',
          status: 'pending',
          has_login_form: false,
          created_at: '2025-01-01T12:00:00Z',
          updated_at: '2025-01-01T12:00:00Z',
        },
      }), 100))
    )
    
    render(<AddUrl />)
    
    const input = screen.getByLabelText(/website url/i)
    const submitButton = screen.getByRole('button', { name: /start web analysis/i })
    
    await user.type(input, 'https://example.com')
    await user.click(submitButton)
    
    // Should show loading state
    expect(screen.getByText('Starting Analysis...')).toBeInTheDocument()
    expect(submitButton).toBeDisabled()
    
    // Wait for completion
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/')
    })
  })

  test('back to dashboard link works', () => {
    render(<AddUrl />)
    
    const backButton = screen.getByText('Back to Dashboard')
    expect(backButton.tagName.toLowerCase()).toBe('button')
  })
}) 