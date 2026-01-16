import { ReactElement, ReactNode } from 'react'
import { render, RenderOptions } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

// Создаем тестовый QueryClient с отключенными retry
const createTestQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0, // Отключаем кэширование для тестов
      },
      mutations: {
        retry: false,
      },
    },
  })

interface AllTheProvidersProps {
  children: ReactNode
}

// Wrapper со всеми необходимыми провайдерами
function AllTheProviders({ children }: AllTheProvidersProps) {
  const testQueryClient = createTestQueryClient()

  return (
    <QueryClientProvider client={testQueryClient}>
      {children}
    </QueryClientProvider>
  )
}

// Кастомный render с провайдерами
function renderWithProviders(
  ui: ReactElement,
  options?: Omit<RenderOptions, 'wrapper'>
) {
  return render(ui, { wrapper: AllTheProviders, ...options })
}

// Re-export everything
export * from '@testing-library/react'
export { renderWithProviders as render }
