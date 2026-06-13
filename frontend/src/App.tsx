import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import AppRoutes from '@/routes/AppRoutes'
import { useApplyTheme } from '@/features/theme/useApplyTheme'

const queryClient = new QueryClient()

function App() {
  useApplyTheme()

  return (
    <QueryClientProvider client={queryClient}>
      <AppRoutes />
    </QueryClientProvider>
  )
}

export default App
