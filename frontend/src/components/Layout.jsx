import React, { useState } from "react"
import { NavLink, useNavigate } from "react-router-dom"
import { useAuth } from "@/hooks/useAuth"
import { Button } from "@/components/ui/button"
import { Users, Calendar, User, TrendingUp, LogOut, LogIn, Menu, X, Zap } from "lucide-react"
import { cn } from "@/lib/utils"

const navItems = [
  { to: "/teams", label: "Teams", Icon: Users },
  { to: "/matches", label: "Matches", Icon: Calendar },
  { to: "/players", label: "Players", Icon: User },
  { to: "/elo", label: "ELO", Icon: TrendingUp },
  { to: "/simulate", label: "Simulate", Icon: Zap },
]

export default function Layout({ children }) {
  const { isAuthenticated, username, signOut } = useAuth()
  const navigate = useNavigate()
  const [mobileOpen, setMobileOpen] = useState(false)

  const handleSignOut = () => {
    signOut()
    navigate("/auth")
  }

  const handleNavClick = () => setMobileOpen(false)

  return (
    <div className="min-h-screen bg-slate-50">
      {/* Top nav */}
      <header className="bg-white border-b border-slate-200 sticky top-0 z-20">
        <div className="max-w-6xl mx-auto px-4 py-3 flex items-center gap-3">
          {/* Logo */}
          <NavLink to="/" className="font-bold text-xl text-slate-900 flex items-center gap-2 shrink-0">
            ⚽ <span className="hidden sm:inline">Football API</span>
          </NavLink>

          {/* Desktop nav */}
          <nav className="hidden md:flex items-center gap-1 flex-1">
            {navItems.map(({ to, label, Icon }) => (
              <NavLink
                key={to}
                to={to}
                className={({ isActive }) =>
                  cn(
                    "flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm font-medium transition-colors",
                    isActive
                      ? "bg-slate-900 text-white"
                      : "text-slate-600 hover:bg-slate-100 hover:text-slate-900"
                  )
                }
              >
                <Icon className="h-4 w-4" />
                {label}
              </NavLink>
            ))}
          </nav>

          {/* Spacer on mobile */}
          <div className="flex-1 md:hidden" />

          {/* Auth area */}
          <div className="hidden md:flex items-center gap-2">
            {isAuthenticated ? (
              <>
                <span className="text-sm text-slate-500 hidden lg:block">👤 {username}</span>
                <Button variant="ghost" size="sm" onClick={handleSignOut}>
                  <LogOut className="h-4 w-4 mr-1" />
                  Sign Out
                </Button>
              </>
            ) : (
              <Button size="sm" onClick={() => navigate("/auth")}>
                <LogIn className="h-4 w-4 mr-1" />
                Sign In
              </Button>
            )}
          </div>

          {/* Mobile hamburger */}
          <button
            className="md:hidden p-2 rounded-md text-slate-600 hover:bg-slate-100"
            onClick={() => setMobileOpen((v) => !v)}
            aria-label="Toggle menu"
          >
            {mobileOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
          </button>
        </div>

        {/* Mobile menu panel */}
        {mobileOpen && (
          <div className="md:hidden border-t border-slate-100 bg-white px-4 py-3 space-y-1">
            {navItems.map(({ to, label, Icon }) => (
              <NavLink
                key={to}
                to={to}
                onClick={handleNavClick}
                className={({ isActive }) =>
                  cn(
                    "flex items-center gap-2 px-3 py-2 rounded-md text-sm font-medium transition-colors w-full",
                    isActive
                      ? "bg-slate-900 text-white"
                      : "text-slate-600 hover:bg-slate-100 hover:text-slate-900"
                  )
                }
              >
                <Icon className="h-4 w-4" />
                {label}
              </NavLink>
            ))}
            <div className="pt-2 border-t border-slate-100 mt-2">
              {isAuthenticated ? (
                <div className="flex items-center justify-between">
                  <span className="text-sm text-slate-500">👤 {username}</span>
                  <Button variant="ghost" size="sm" onClick={() => { handleSignOut(); handleNavClick() }}>
                    <LogOut className="h-4 w-4 mr-1" />
                    Sign Out
                  </Button>
                </div>
              ) : (
                <Button size="sm" className="w-full" onClick={() => { navigate("/auth"); handleNavClick() }}>
                  <LogIn className="h-4 w-4 mr-1" />
                  Sign In
                </Button>
              )}
            </div>
          </div>
        )}
      </header>

      {/* Main content */}
      <main className="max-w-6xl mx-auto px-4 py-8">
        {children}
      </main>
    </div>
  )
}
