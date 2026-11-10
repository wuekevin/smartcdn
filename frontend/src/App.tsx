import { useEffect, useState } from 'react';
import { BrowserRouter, Routes, Route, Navigate, useNavigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Spin } from '@arco-design/web-react';
import axios from 'axios';

import SetupWizard from './pages/setup/SetupWizard';
import LoginPage from './pages/auth/LoginPage';
import MainLayout from './layouts/MainLayout';
import DashboardPage from './pages/dashboard/DashboardPage';

import DNSAccountsPage from './pages/dns/DNSAccountsPage';
import DomainsPage from './pages/dns/DomainsPage';
import DNSRecordsPage from './pages/dns/DNSRecordsPage';
import PublishJobsPage from './pages/dns/PublishJobsPage';

import CDNProvidersPage from './pages/cdn/CDNProvidersPage';
import CDNSitesPage from './pages/cdn/CDNSitesPage';

import NodesPage from './pages/scheduler/NodesPage';
import NodePoolsPage from './pages/scheduler/NodePoolsPage';
import ProbesPage from './pages/scheduler/ProbesPage';
import CandidatesPage from './pages/scheduler/CandidatesPage';

import CertificatesPage from './pages/ssl/CertificatesPage';
import WorkflowsPage from './pages/ssl/WorkflowsPage';

import MetricsPage from './pages/monitor/MetricsPage';
import AlertsPage from './pages/monitor/AlertsPage';
import AlertChannelsPage from './pages/monitor/AlertChannelsPage';

import UsersPage from './pages/system/UsersPage';
import AuditPage from './pages/system/AuditPage';

const queryClient = new QueryClient({
  defaultOptions: { queries: { staleTime: 10_000, retry: 1 } },
});

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const token = localStorage.getItem('token');
  if (!token) return <Navigate to="/login" replace />;
  return <>{children}</>;
}

function AppInitializer() {
  const [checking, setChecking] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    axios.get('/api/v1/setup/check').then(() => {
      navigate('/setup', { replace: true });
    }).catch((err) => {
      if (err?.response?.status !== 403) {
        // unknown error — stay put
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
          <Route path="/" element={<AppInitializer />} />
          <Route path="/setup" element={<SetupWizard />} />
          <Route path="/login" element={<LoginPage />} />

          <Route path="/" element={<MainLayout />}>
            <Route path="dashboard" element={<ProtectedRoute><DashboardPage /></ProtectedRoute>} />

            <Route path="dns/accounts" element={<ProtectedRoute><DNSAccountsPage /></ProtectedRoute>} />
            <Route path="dns/domains" element={<ProtectedRoute><DomainsPage /></ProtectedRoute>} />
            <Route path="dns/records" element={<ProtectedRoute><DNSRecordsPage /></ProtectedRoute>} />
            <Route path="dns/publish-jobs" element={<ProtectedRoute><PublishJobsPage /></ProtectedRoute>} />

            <Route path="cdn/providers" element={<ProtectedRoute><CDNProvidersPage /></ProtectedRoute>} />
            <Route path="cdn/sites" element={<ProtectedRoute><CDNSitesPage /></ProtectedRoute>} />

            <Route path="scheduler/nodes" element={<ProtectedRoute><NodesPage /></ProtectedRoute>} />
            <Route path="scheduler/node-pools" element={<ProtectedRoute><NodePoolsPage /></ProtectedRoute>} />
            <Route path="scheduler/probes" element={<ProtectedRoute><ProbesPage /></ProtectedRoute>} />
            <Route path="scheduler/candidates" element={<ProtectedRoute><CandidatesPage /></ProtectedRoute>} />

            <Route path="ssl/certificates" element={<ProtectedRoute><CertificatesPage /></ProtectedRoute>} />
            <Route path="ssl/workflows" element={<ProtectedRoute><WorkflowsPage /></ProtectedRoute>} />

            <Route path="monitor/metrics" element={<ProtectedRoute><MetricsPage /></ProtectedRoute>} />
            <Route path="monitor/alerts" element={<ProtectedRoute><AlertsPage /></ProtectedRoute>} />
            <Route path="monitor/channels" element={<ProtectedRoute><AlertChannelsPage /></ProtectedRoute>} />

            <Route path="system/users" element={<ProtectedRoute><UsersPage /></ProtectedRoute>} />
            <Route path="system/audit" element={<ProtectedRoute><AuditPage /></ProtectedRoute>} />

            <Route path="*" element={<Navigate to="/dashboard" replace />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
