/* eslint-disable react-refresh/only-export-components */
import React, { createContext, useContext, useState } from "react"

const AuthContext = createContext(null)

export function AuthProvider({ children }) {
  const [token, setToken] = useState(() => localStorage.getItem("token"))
  const [username, setUsername] = useState(() => localStorage.getItem("username"))

  const signIn = (token, username) => {
    localStorage.setItem("token", token)
    localStorage.setItem("username", username)
    setToken(token)
    setUsername(username)
  }

  const signOut = () => {
    localStorage.removeItem("token")
    localStorage.removeItem("username")
    setToken(null)
    setUsername(null)
  }

  return (
    <AuthContext.Provider value={{ token, username, isAuthenticated: !!token, signIn, signOut }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  return useContext(AuthContext)
}
