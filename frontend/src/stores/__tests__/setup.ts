import { vi } from 'vitest'
import { defineStore } from 'pinia'

// Mock Wails runtime
vi.mock('../../../wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn(),
  EventsOff: vi.fn(),
  EventsEmit: vi.fn()
}))

// Mock projectStore
vi.mock('../projectStore', () => ({
  useProjectStore: vi.fn(() => ({
    projects: [
      { path: '/test/project1', name: 'Project 1' },
      { path: '/test/project2', name: 'Project 2' },
      { path: '/test/project3', name: 'Project 3' }
    ]
  }))
}))

// Mock window object properties if needed
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
})
