import React from "react"
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom"
import { AuthProvider } from "@/hooks/useAuth"
import Layout from "@/components/Layout"
import HomePage from "@/pages/HomePage"
import AuthPage from "@/pages/AuthPage"
import TeamsPage from "@/pages/TeamsPage"
import MatchesPage from "@/pages/MatchesPage"
import MatchDetailPage from "@/pages/MatchDetailPage"
import PlayersPage from "@/pages/PlayersPage"
import EloPage from "@/pages/EloPage"

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/auth" element={<AuthPage />} />
          <Route
            path="/*"
            element={
              <Layout>
                <Routes>
                  <Route path="/" element={<HomePage />} />
                  <Route path="/teams" element={<TeamsPage />} />
                  <Route path="/matches" element={<MatchesPage />} />
                  <Route path="/matches/:id" element={<MatchDetailPage />} />
                  <Route path="/players" element={<PlayersPage />} />
                  <Route path="/elo" element={<EloPage />} />
                  <Route path="*" element={<Navigate to="/" replace />} />
                </Routes>
              </Layout>
            }
          />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  )
}
