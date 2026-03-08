import React, { useState, useEffect } from "react"
import { useAuth } from "@/hooks/useAuth"
import { getTeams, simulateMatch } from "@/api/client"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Lock } from "lucide-react"
import { useNavigate } from "react-router-dom"

const DEFAULT_SIMULATIONS = 1000

function PctBar({ label, pct, colorClass }) {
  return (
    <div className="space-y-1">
      <div className="flex justify-between text-sm">
        <span className="font-medium">{label}</span>
        <span className="tabular-nums font-semibold">{(pct * 100).toFixed(1)}%</span>
      </div>
      <div className="h-3 w-full rounded-full bg-slate-100 overflow-hidden">
        <div
          className={`h-full rounded-full transition-all ${colorClass}`}
          style={{ width: `${(pct * 100).toFixed(1)}%` }}
        />
      </div>
    </div>
  )
}

function StatCard({ label, value, sub }) {
  return (
    <div className="text-center p-3 rounded-lg bg-slate-50 border">
      <p className="text-2xl font-black tabular-nums">{value}</p>
      <p className="text-xs font-semibold text-slate-600 mt-0.5">{label}</p>
      {sub && <p className="text-xs text-slate-400 mt-0.5">{sub}</p>}
    </div>
  )
}

export default function SimulatePage() {
  const { isAuthenticated } = useAuth()
  const navigate = useNavigate()

  const [teams, setTeams] = useState([])
  const [homeTeamId, setHomeTeamId] = useState("")
  const [awayTeamId, setAwayTeamId] = useState("")
  const [date, setDate] = useState("")
  const [venue, setVenue] = useState("neutral")
  const [simulations, setSimulations] = useState(DEFAULT_SIMULATIONS)

  const [result, setResult] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState("")

  useEffect(() => {
    getTeams()
      .then((r) => setTeams(r.data?.data || []))
      .catch(() => {})
  }, [])

  const handleSubmit = async (e) => {
    e.preventDefault()
    if (!homeTeamId || !awayTeamId) return
    setLoading(true)
    setError("")
    setResult(null)
    try {
      const body = {
        homeTeamId: Number(homeTeamId),
        awayTeamId: Number(awayTeamId),
        venue,
        simulations: Number(simulations),
      }
      if (date) body.date = date
      const res = await simulateMatch(body)
      setResult(res.data)
    } catch (err) {
      setError(err.response?.data?.error || "Simulation failed. Please try again.")
    } finally {
      setLoading(false)
    }
  }

  if (!isAuthenticated) {
    return (
      <div className="flex flex-col items-center justify-center py-24 space-y-4 text-center">
        <Lock className="h-12 w-12 text-slate-300" />
        <h2 className="text-2xl font-bold text-slate-700">Sign in required</h2>
        <p className="text-slate-500 max-w-sm">
          The match simulator requires an account. Please sign in to access this feature.
        </p>
        <Button onClick={() => navigate("/auth")}>Sign In</Button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <h2 className="text-2xl font-bold">Match Simulator</h2>

      {/* Form */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Simulation Parameters</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="grid gap-4 sm:grid-cols-2">
              {/* Home team */}
              <div className="space-y-1">
                <Label>Home Team</Label>
                <select
                  className="flex h-10 w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-950"
                  value={homeTeamId}
                  onChange={(e) => setHomeTeamId(e.target.value)}
                  required
                >
                  <option value="">Select home team…</option>
                  {teams.map((t) => (
                    <option key={t.id} value={t.id}>{t.name}</option>
                  ))}
                </select>
              </div>

              {/* Away team */}
              <div className="space-y-1">
                <Label>Away Team</Label>
                <select
                  className="flex h-10 w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-950"
                  value={awayTeamId}
                  onChange={(e) => setAwayTeamId(e.target.value)}
                  required
                >
                  <option value="">Select away team…</option>
                  {teams.map((t) => (
                    <option key={t.id} value={t.id}>{t.name}</option>
                  ))}
                </select>
              </div>
            </div>

            <div className="grid gap-4 sm:grid-cols-3">
              {/* Venue */}
              <div className="space-y-1">
                <Label>Venue</Label>
                <select
                  className="flex h-10 w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-950"
                  value={venue}
                  onChange={(e) => setVenue(e.target.value)}
                >
                  <option value="neutral">Neutral</option>
                  <option value="home">Home team's ground</option>
                  <option value="away">Away team's ground</option>
                </select>
              </div>

              {/* Date */}
              <div className="space-y-1">
                <Label>As-of Date (optional)</Label>
                <Input
                  type="date"
                  value={date}
                  onChange={(e) => setDate(e.target.value)}
                  className="w-full"
                />
              </div>

              {/* Simulations */}
              <div className="space-y-1">
                <Label>Simulations (1–10,000)</Label>
                <Input
                  type="number"
                  min={1}
                  max={10000}
                  value={simulations}
                  onChange={(e) => setSimulations(e.target.value)}
                  className="w-full"
                />
              </div>
            </div>

            <div className="flex justify-end">
              <Button type="submit" disabled={loading} className="min-w-[120px]">
                {loading ? "Simulating…" : "Run Simulation"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>

      {/* Error */}
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* Results */}
      {result && (
        <div className="space-y-4">
          {/* Match header */}
          <Card>
            <CardContent className="p-4">
              <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
                <div className="text-center flex-1">
                  <p className="text-xl font-bold">{result.homeTeam}</p>
                  <p className="text-sm text-slate-500 mt-0.5">Home · ELO {Math.round(result.homeElo)}</p>
                </div>
                <div className="text-center px-4">
                  <Badge variant="outline" className="text-xs uppercase tracking-wide">
                    {result.venue}
                  </Badge>
                  <p className="text-xs text-slate-400 mt-1">as of {result.asOf}</p>
                  <p className="text-xs text-slate-400">{result.simulations.toLocaleString()} iterations</p>
                </div>
                <div className="text-center flex-1">
                  <p className="text-xl font-bold">{result.awayTeam}</p>
                  <p className="text-sm text-slate-500 mt-0.5">Away · ELO {Math.round(result.awayElo)}</p>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Outcome probabilities */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Outcome Probabilities</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <PctBar
                label={`${result.homeTeam} Win`}
                pct={result.outcome.homeWinPct}
                colorClass="bg-blue-500"
              />
              <PctBar
                label="Draw"
                pct={result.outcome.drawPct}
                colorClass="bg-slate-400"
              />
              <PctBar
                label={`${result.awayTeam} Win`}
                pct={result.outcome.awayWinPct}
                colorClass="bg-rose-500"
              />

              {/* Confidence intervals */}
              <div className="mt-2 rounded-lg bg-slate-50 border p-3 text-xs text-slate-500 space-y-1">
                <p className="font-semibold text-slate-600 mb-1">95% Wilson Confidence Intervals</p>
                <p>
                  <span className="font-medium text-blue-600">{result.homeTeam} Win:</span>{" "}
                  {(result.outcome.homeWinCI[0] * 100).toFixed(1)}% – {(result.outcome.homeWinCI[1] * 100).toFixed(1)}%
                </p>
                <p>
                  <span className="font-medium text-slate-600">Draw:</span>{" "}
                  {(result.outcome.drawCI[0] * 100).toFixed(1)}% – {(result.outcome.drawCI[1] * 100).toFixed(1)}%
                </p>
                <p>
                  <span className="font-medium text-rose-600">{result.awayTeam} Win:</span>{" "}
                  {(result.outcome.awayWinCI[0] * 100).toFixed(1)}% – {(result.outcome.awayWinCI[1] * 100).toFixed(1)}%
                </p>
              </div>
            </CardContent>
          </Card>

          {/* Stats grid */}
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            <StatCard
              label="Expected Home Goals"
              value={result.expectedScore.homeGoals.toFixed(2)}
              sub={result.homeTeam}
            />
            <StatCard
              label="Expected Away Goals"
              value={result.expectedScore.awayGoals.toFixed(2)}
              sub={result.awayTeam}
            />
            <StatCard
              label="Upset Probability"
              value={`${(result.upsetProbability * 100).toFixed(1)}%`}
              sub="Lower-rated team wins"
            />
            <StatCard
              label="Simulations Run"
              value={result.simulations.toLocaleString()}
            />
          </div>
        </div>
      )}
    </div>
  )
}
