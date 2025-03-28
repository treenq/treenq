import './App.css';

import { Layout } from '@/Layout';
import { SignInPage } from '@/pages/SignIn';
import { BrowserRouter, Route, Routes, Navigate, useLocation } from 'react-router';
import { useAuthStore } from '@/stores/useAuthStore'

type ProtectedRageProps = {
  children: React.ReactNode

  redirect: string
  satisfies: boolean
}

function ProtectedPage({children, redirect, satisfies}: ProtectedRageProps) {
  if (!satisfies) return <Navigate to={redirect} replace />

    return children;
}

function AppRoutes() {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route
            index
            element={
              <ProtectedPage redirect='signIn' satisfies={isAuthenticated}>
                <div>Protected Home Page</div>
              </ProtectedPage>
            }
          />
          <Route path="signIn" element={
            <ProtectedPage redirect='/' satisfies={!isAuthenticated}>
              <SignInPage />
            </ProtectedPage>
          } />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

function App() {
  return (
      <AppRoutes />
  );
}

export default App;
