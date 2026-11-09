import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import SetupWizard from './pages/setup/SetupWizard';

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/setup" element={<SetupWizard />} />
        <Route path="/" element={<Navigate to="/setup" replace />} />
      </Routes>
    </BrowserRouter>
  );
}
