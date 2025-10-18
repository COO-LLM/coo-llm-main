import React from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { ThemeProvider } from '@/contexts/theme-context'
import { AuthProvider, useAuth } from '@/contexts/auth-context'
import { DashboardLayout } from '@/components/dashboard-layout'
import LoginPage from '@/app/login/page'
import Dashboard from '@/app/page'
import MetricsPage from '@/app/metrics/page'
import ClientsPage from '@/app/clients/page'
import ConfigurationPage from '@/app/configuration/page'
import StatisticsPage from '@/app/statistics/page'

function AppRoutes() {
  const { user, isLoading } = useAuth()

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-primary"></div>
      </div>
    )
  }

  return (
    <Routes>
      {!user ? (
        <>
          <Route path="/login" element={<LoginPage />} />
          <Route path="*" element={<Navigate to="/login" replace />} />
        </>
      ) : (
        <>
          <Route
            path="/"
            element={
              <DashboardLayout>
                <Dashboard />
              </DashboardLayout>
            }
          />
          <Route
            path="/metrics"
            element={
              <DashboardLayout>
                <MetricsPage />
              </DashboardLayout>
            }
          />
          <Route
            path="/clients"
            element={
              <DashboardLayout>
                <ClientsPage />
              </DashboardLayout>
            }
          />
          <Route
            path="/configuration"
            element={
              <DashboardLayout>
                <ConfigurationPage />
              </DashboardLayout>
            }
          />
          <Route
            path="/statistics"
            element={
              <DashboardLayout>
                <StatisticsPage />
              </DashboardLayout>
            }
          />
          <Route path="*" element={<Navigate to="/" replace />} />
        </>
      )}
    </Routes>
  )
}

export default function App() {
  return (
    <ThemeProvider>
      <AuthProvider>
        <BrowserRouter basename="/ui">
          <AppRoutes />
        </BrowserRouter>
      </AuthProvider>
    </ThemeProvider>
  )
}
