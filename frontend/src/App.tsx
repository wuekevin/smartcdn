import { useEffect, useState } from 'react';
import { BrowserRouter, Routes, Route, Navigate, useNavigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Spin } from '@arco-design/web-react';
import axios from 'axios';
import SetupWizard from './pages/setup/SetupWizard';
import LoginPage from './pages/auth/LoginPage';
import MainLayout from './layouts/MainLayout';
import DashboardPage from './pages/dashboard/DashboardPage';
import DomainsPage from './pages/dns/DomainsPage';
import CertificatesPage from './pages/ssl/CertificatesPage';
import AlertsPage from './pages/monitor/AlertsPage';
import NodesPage from './pages/scheduler/NodesPage';

const queryClient = new QueryClient({
  defaultOptions: { queries: { staleTime: 10_000, retry: 1 } },
});

// Checks whether install.lock exists and redirects to /setup if not.
function AppInitializer() {
  const [checking, setChecking] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    axios.get('/api/v1/setup/check').then(() => {
      // 200 means not yet installed
      navigate('/setup', { replace: true });
    }).catch((err) => {
      if (err?.response?.status === 403) {
        // 403 SETUP_LOCKED = already installed, stay on main app
      } else {
        // Unknown error; stay put and let login/dashboard handle it
      }
    }).finally(() => setChecking(false));
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  if (checking) {
    return <Spin style={{ display: 'block', margin: '80px auto' }} />;
  }
  return <Navigate to="/dashboard" replace />;
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          {/* Root — detect install state then redirect */}
          <Route path="/" element={<AppInitializer />} />

          {/* Setup wizard */}
          <Route path="/setup" element={<SetupWizard />} />

          {/* Login */}
          <Route path="/login" element={<LoginPage />} />

          {/* Main authenticated app */}
          <Route path="/" element={<MainLayout />}>
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
