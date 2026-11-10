import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import SetupWizard from './pages/setup/SetupWizard';
import MainLayout from './layouts/MainLayout';
import DashboardPage from './pages/dashboard/DashboardPage';
import DomainsPage from './pages/dns/DomainsPage';
import CertificatesPage from './pages/ssl/CertificatesPage';
import AlertsPage from './pages/monitor/AlertsPage';
import NodesPage from './pages/scheduler/NodesPage';

const queryClient = new QueryClient({
  defaultOptions: { queries: { staleTime: 10_000, retry: 1 } },
});

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          {/* Setup wizard — shown before install completes */}
          <Route path="/setup" element={<SetupWizard />} />

          {/* Main authenticated app */}
          <Route path="/" element={<MainLayout />}>
            <Route index element={<Navigate to="/dashboard" replace />} />
            <Route path="dashboard" element={<DashboardPage />} />

            {/* DNS */}
            <Route path="dns/domains" element={<DomainsPage />} />

            {/* SSL */}
            <Route path="ssl/certificates" element={<CertificatesPage />} />

            {/* Scheduler */}
            <Route path="scheduler/nodes" element={<NodesPage />} />

            {/* Monitor */}
            <Route path="monitor/alerts" element={<AlertsPage />} />

            {/* Catch-all */}
            <Route path="*" element={<Navigate to="/dashboard" replace />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
