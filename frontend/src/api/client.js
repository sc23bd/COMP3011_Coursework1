import axios from "axios"

const API_BASE = "/api/v1"

const api = axios.create({
  baseURL: API_BASE,
  headers: { "Content-Type": "application/json" },
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem("token")
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Auth
export const login = (data) => api.post("/auth/login", data)
export const register = (data) => api.post("/auth/register", data)

// Teams
export const getTeams = () => api.get("/football/teams")
export const getTeam = (id) => api.get(`/football/teams/${id}`)
export const createTeam = (data) => api.post("/football/teams", data)
export const updateTeam = (id, data) => api.put(`/football/teams/${id}`, data)
export const deleteTeam = (id) => api.delete(`/football/teams/${id}`)
export const getTeamHistory = (id) => api.get(`/football/teams/${id}/history`)

// Matches
export const getMatches = (params) => api.get("/football/matches", { params })
export const getMatch = (id) => api.get(`/football/matches/${id}`)
export const getHeadToHead = (teamA, teamB) =>
  api.get("/football/head-to-head", { params: { teamA, teamB } })

// Goals
export const getGoals = (matchId) => api.get(`/football/matches/${matchId}/goals`)
export const createGoal = (matchId, data) => api.post(`/football/matches/${matchId}/goals`, data)
export const deleteGoal = (matchId, goalId) =>
  api.delete(`/football/matches/${matchId}/goals/${goalId}`)

// Shootouts
export const getShootout = (matchId) => api.get(`/football/matches/${matchId}/shootout`)
export const createShootout = (matchId, data) =>
  api.post(`/football/matches/${matchId}/shootout`, data)
export const deleteShootout = (matchId) => api.delete(`/football/matches/${matchId}/shootout`)

// Players
export const getPlayerGoals = (name) => api.get(`/football/players/${encodeURIComponent(name)}/goals`)

// ELO
export const getEloRankings = (params) => api.get("/football/rankings/elo", { params })
export const recalculateElo = (params) => api.post("/football/rankings/elo/recalculate", null, { params })
export const getTeamElo = (id, params) => api.get(`/football/teams/${id}/elo`, { params })
export const getTeamEloTimeline = (id, params) =>
  api.get(`/football/teams/${id}/elo/timeline`, { params })

export default api
