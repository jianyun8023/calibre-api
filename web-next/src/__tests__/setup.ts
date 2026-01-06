/**
 * Test Setup
 * 
 * This file sets up the testing environment for all tests.
 */

import '@testing-library/jest-dom'

// Mock fetch if not available
if (typeof global.fetch === 'undefined') {
  global.fetch = jest.fn() as any
}

// Mock EventSource for SSE tests
if (typeof global.EventSource === 'undefined') {
  global.EventSource = class MockEventSource {
    constructor(public url: string) {}
    addEventListener = jest.fn()
    removeEventListener = jest.fn()
    close = jest.fn()
    dispatchEvent = jest.fn()
    onopen: any = null
    onmessage: any = null
    onerror: any = null
    readyState = 0
    CONNECTING = 0
    OPEN = 1
    CLOSED = 2
  } as any
}

// Mock localStorage
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
  length: 0,
  key: jest.fn(),
}

global.localStorage = localStorageMock as any

// Reset mocks before each test
beforeEach(() => {
  jest.clearAllMocks()
  localStorageMock.getItem.mockClear()
  localStorageMock.setItem.mockClear()
  localStorageMock.removeItem.mockClear()
  localStorageMock.clear.mockClear()
})

