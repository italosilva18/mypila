import React from 'react';
import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import { Dashboard } from './components/Dashboard';
import { CompanyList } from './components/CompanyList';
import { Layout } from './components/Layout';
import { Categories } from './pages/Categories';
import { Reports } from './pages/Reports';
import { Recurring } from './pages/Recurring';
import { Login } from './pages/Login';
import { Register } from './pages/Register';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { ErrorBoundary } from './components/ErrorBoundary';
import { Loader2 } from 'lucide-react';

const PrivateRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { user, loading } = useAuth();
  const location = useLocation();

  if (loading) {
    return (
      <div className="min-h-screen bg-stone-50 flex items-center justify-center">
        <Loader2 className="w-6 h-6 md:w-8 md:h-8 text-stone-600 animate-spin" />
      </div>
    );
  }

  if (!user) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  return <>{children}</>;
};

const AppRoutes: React.FC = () => {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />

      <Route path="/" element={
        <PrivateRoute>
          <CompanyList />
        </PrivateRoute>
      } />

      <Route path="/company/:companyId" element={
        <PrivateRoute>
          <Layout />
        </PrivateRoute>
      }>
        <Route path="dashboard" element={<Dashboard />} />
        <Route path="categories" element={<Categories />} />
        <Route path="reports" element={<Reports />} />
        <Route path="recurring" element={<Recurring />} />
        {/* Redirect root company path to dashboard */}
        <Route index element={<Navigate to="dashboard" replace />} />
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

import { DateFilterProvider } from './contexts/DateFilterContext';
import { ToastProvider } from './contexts/ToastContext';
import { ToastContainer } from './components/ToastContainer';

const App: React.FC = () => {
  return (
    <ErrorBoundary>
      <BrowserRouter>
        <AuthProvider>
          <DateFilterProvider>
            <ToastProvider>
              <AppRoutes />
              <ToastContainer />
            </ToastProvider>
          </DateFilterProvider>
        </AuthProvider>
      </BrowserRouter>
    </ErrorBoundary>
  );
};

export default App;
