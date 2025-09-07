import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import App from './App.tsx'
import { QueryClientProvider } from '@tanstack/react-query'
import CssBaseline from '@mui/material/CssBaseline';
import queryClient from './services/queryClient.ts';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <CssBaseline />
        <App />
    </QueryClientProvider>
  </StrictMode>,
)
