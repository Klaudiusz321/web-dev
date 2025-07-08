describe('Basic Jest Test', () => {
  test('Jest is working', () => {
    expect(1 + 1).toBe(2)
  })

  test('Mock function works', () => {
    const mockFn = jest.fn()
    mockFn('test')
    expect(mockFn).toHaveBeenCalledWith('test')
  })
}) 