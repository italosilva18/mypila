import React, { Suspense } from 'react';
import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { DateFilterProvider } from './contexts/DateFilterContext';
import { ToastProvider } from './contexts/ToastContext';
import { ErrorBoundary } from './components/ErrorBoundary';
import { ToastContainer } from './components/ToastContainer';
import { Layout } from './components/Layout';
import { Loader2 } from 'lucide-react';

// Lazy load pages for better performance (code splitting)
const Dashboard = React.lazy(() => import('./components/Dashboard').then(m => ({ default: m.Dashboard })));
const CompanyList = React.lazy(() => import('./components/CompanyList').then(m => ({ default: m.CompanyList })));
const Categories = React.lazy(() => import('./pages/Categories').then(m => ({ default: m.Categories })));
const Reports = React.lazy(() => import('./pages/Reports').then(m => ({ default: m.Reports })));
const Recurring = React.lazy(() => import('./pages/Recurring').then(m => ({ default: m.Recurring })));
const Quotes = React.lazy(() => import('./pages/Quotes').then(m => ({ default: m.Quotes })));
const QuoteComparisonPage = React.lazy(() => import('./pages/QuoteComparison').then(m => ({ default: m.QuoteComparisonPage })));
const Login = React.lazy(() => import('./pages/Login').then(m => ({ default: m.Login })));
const Register = React.lazy(() => import('./pages/Register').then(m => ({ default: m.Register })));

// Loading fallback component
const PageLoader: React.FC = () => (
  <div className="min-h-screen bg-stone-50 flex items-center justify-center">
    <Loader2 className="w-6 h-6 md:w-8 md:h-8 text-stone-600 animate-spin" />
  </div>
);

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
    <Suspense fallback={<PageLoader />}>
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
          <Route path="quotes" element={<Quotes />} />
          <Route path="quotes/:quoteId/comparison" element={<QuoteComparisonPage />} />
          {/* Redirect root company path to dashboard */}
          <Route index element={<Navigate to="dashboard" replace />} />
        </Route>

        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </Suspense>
  );
};

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
