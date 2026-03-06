import React, { useState, useEffect, useCallback } from "react"
import { useNavigate } from "react-router-dom"
import { useAuth } from "@/hooks/useAuth"
import { getEloRankings, getTeamElo, getTeamEloTimeline, getTeams, recalculateElo } from "@/api/client"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { TrendingUp, TrendingDown, Minus, RefreshCw, ChevronLeft, ChevronRight } from "lucide-react"

function RankBadge({ rank }) {
  const colors = rank === 1 ? "bg-yellow-100 text-yellow-800" : rank === 2 ? "bg-slate-100 text-slate-700" : rank === 3 ? "bg-orange-100 text-orange-800" : "bg-white text-slate-600 border"
  return <span className={`inline-flex items-center justify-center w-8 h-8 rounded-full text-sm font-bold ${colors}`}>{rank}</span>
}

export default function EloPage() {
  const { isAuthenticated } = useAuth()
  const navigate = useNavigate()

  // Rankings tab
  const [rankings, setRankings] = useState([])
  const [rankingsLoading, setRankingsLoading] = useState(true)
  const [rankingsError, setRankingsError] = useState("")
  const [rankDate, setRankDate] = useState("")
  const [rankRegion, setRankRegion] = useState("")
  const [rankLimit] = useState(50)
  const [rankOffset, setRankOffset] = useState(0)

  // Team ELO tab
  const [teams, setTeams] = useState([])
  const [selectedTeam, setSelectedTeam] = useState("")
  const [eloDate, setEloDate] = useState("")
  const [teamElo, setTeamElo] = useState(null)
  const [eloLoading, setEloLoading] = useState(false)
  const [eloError, setEloError] = useState("")

  // Timeline tab
  const [tlTeam, setTlTeam] = useState("")
  const [tlStart, setTlStart] = useState("")
  const [tlEnd, setTlEnd] = useState("")
  const [tlResolution, setTlResolution] = useState("match")
  const [timeline, setTimeline] = useState(null)
  const [tlLoading, setTlLoading] = useState(false)
  const [tlError, setTlError] = useState("")

  // Recalculate
  const [recalcMsg, setRecalcMsg] = useState("")

  const loadRankings = useCallback(async () => {
    setRankingsLoading(true)
    setRankingsError("")
    try {
      const params = { limit: rankLimit, offset: rankOffset }
      if (rankDate) params.date = rankDate
      if (rankRegion) params.region = rankRegion
      const res = await getEloRankings(params)
      setRankings(res.data?.data || [])
    } catch {
      setRankingsError("Failed to load rankings")
    } finally {
      setRankingsLoading(false)
    }
  }, [rankDate, rankRegion, rankLimit, rankOffset])

  useEffect(() => { loadRankings() }, [loadRankings])
  useEffect(() => { getTeams().then((r) => setTeams(r.data?.data || [])).catch(() => {}) }, [])

  const handleTeamElo = async (e) => {
    e.preventDefault()
    if (!selectedTeam) return
    setEloLoading(true)
    setEloError("")
    setTeamElo(null)
    try {
      const params = {}
      if (eloDate) params.date = eloDate
      const res = await getTeamElo(selectedTeam, params)
      setTeamElo(res.data)
    } catch (err) {
      setEloError(err.response?.data?.error || "Failed to load ELO rating")
    } finally {
      setEloLoading(false)
    }
  }

  const handleTimeline = async (e) => {
    e.preventDefault()
    if (!tlTeam) return
    setTlLoading(true)
    setTlError("")
    setTimeline(null)
    try {
      const params = { resolution: tlResolution }
      if (tlStart) params.start_date = tlStart
      if (tlEnd) params.end_date = tlEnd
      const res = await getTeamEloTimeline(tlTeam, params)
      setTimeline(res.data)
    } catch (err) {
      setTlError(err.response?.data?.error || "Failed to load timeline")
    } finally {
      setTlLoading(false)
    }
  }

  const handleRecalculate = async () => {
    setRecalcMsg("")
    try {
      const res = await recalculateElo()
      setRecalcMsg(res.data?.message || "Recalculation started")
    } catch (err) {
      setRecalcMsg(err.response?.data?.error || "Recalculation failed")
    }
  }

  return (
    <div className="space-y-4">
      <h2 className="text-2xl font-bold">ELO Rankings</h2>

      <Tabs defaultValue="rankings">
        <TabsList>
          <TabsTrigger value="rankings">Global Rankings</TabsTrigger>
          <TabsTrigger value="team">Team Rating</TabsTrigger>
          <TabsTrigger value="timeline">ELO Timeline</TabsTrigger>
        </TabsList>

        {/* RANKINGS */}
        <TabsContent value="rankings" className="space-y-4 mt-4">
          <Card>
            <CardContent className="p-4">
              <div className="flex flex-wrap gap-3 items-end">
                <div className="space-y-1">
                  <Label>Date (YYYY-MM-DD)</Label>
                  <Input type="date" value={rankDate} onChange={(e) => { setRankDate(e.target.value); setRankOffset(0) }} className="w-40" />
                </div>
                <div className="space-y-1">
                  <Label>Region</Label>
                  <Input value={rankRegion} onChange={(e) => { setRankRegion(e.target.value); setRankOffset(0) }} placeholder="e.g. europe" className="w-32" />
                </div>
                <Button variant="outline" onClick={loadRankings} size="sm">
                  <RefreshCw className="h-4 w-4 mr-1" /> Refresh
                </Button>
                {isAuthenticated && (
                  <Button variant="secondary" size="sm" onClick={handleRecalculate}>
                    Recalculate ELO
                  </Button>
                )}
              </div>
              {recalcMsg && <p className="text-sm text-green-700 mt-2">{recalcMsg}</p>}
            </CardContent>
          </Card>

          {rankingsError && <Alert variant="destructive"><AlertDescription>{rankingsError}</AlertDescription></Alert>}

          {rankingsLoading ? (
            <p className="text-slate-500 text-center py-8">Loading rankings...</p>
          ) : (
            <>
              <div className="overflow-x-auto rounded-lg border">
                <table className="w-full text-sm">
                  <thead className="bg-slate-50 border-b">
                    <tr>
                      <th className="text-left p-3 font-semibold">Rank</th>
                      <th className="text-left p-3 font-semibold">Team</th>
                      <th className="text-right p-3 font-semibold">ELO Rating</th>
                    </tr>
                  </thead>
                  <tbody>
                    {rankings.map((r) => (
                      <tr key={r.teamId} className="border-b hover:bg-slate-50 transition-colors">
                        <td className="p-3"><RankBadge rank={r.rank} /></td>
                        <td className="p-3 font-medium">{r.teamName}</td>
                        <td className="p-3 text-right tabular-nums font-semibold">{Math.round(r.elo)}</td>
                      </tr>
                    ))}
                    {rankings.length === 0 && (
                      <tr><td colSpan={3} className="p-8 text-center text-slate-500">No ranking data available.</td></tr>
                    )}
                  </tbody>
                </table>
              </div>
              <div className="flex items-center justify-between">
                <Button variant="outline" size="sm" onClick={() => setRankOffset(Math.max(0, rankOffset - rankLimit))} disabled={rankOffset === 0}>
                  <ChevronLeft className="h-4 w-4 mr-1" /> Previous
                </Button>
                <span className="text-sm text-slate-500">Showing {rankOffset + 1}–{rankOffset + rankings.length}</span>
                <Button variant="outline" size="sm" onClick={() => setRankOffset(rankOffset + rankLimit)} disabled={rankings.length < rankLimit}>
                  Next <ChevronRight className="h-4 w-4 ml-1" />
                </Button>
              </div>
            </>
          )}
        </TabsContent>

        {/* TEAM ELO */}
        <TabsContent value="team" className="space-y-4 mt-4">
          <Card>
            <CardContent className="p-4">
              <form onSubmit={handleTeamElo} className="flex flex-wrap gap-3 items-end">
                <div className="space-y-1 flex-1 min-w-[160px]">
                  <Label>Team</Label>
                  <select
                    className="flex h-10 w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-950"
                    value={selectedTeam}
                    onChange={(e) => setSelectedTeam(e.target.value)}
                    required
                  >
                    <option value="">Select team...</option>
                    {teams.map((t) => <option key={t.id} value={t.id}>{t.name}</option>)}
                  </select>
                </div>
                <div className="space-y-1">
                  <Label>Date (optional)</Label>
                  <Input type="date" value={eloDate} onChange={(e) => setEloDate(e.target.value)} className="w-40" />
                </div>
                <Button type="submit" disabled={eloLoading}>
                  {eloLoading ? "Loading..." : "Get Rating"}
                </Button>
              </form>
            </CardContent>
          </Card>

          {eloError && <Alert variant="destructive"><AlertDescription>{eloError}</AlertDescription></Alert>}

          {teamElo && (
            <Card>
              <CardHeader><CardTitle>{teamElo.teamName}</CardTitle></CardHeader>
              <CardContent className="grid grid-cols-2 gap-4 sm:grid-cols-4">
                <div className="text-center">
                  <p className="text-3xl font-black">{Math.round(teamElo.elo)}</p>
                  <p className="text-xs text-slate-500 mt-1">ELO Rating</p>
                </div>
                <div className="text-center">
                  <p className="text-3xl font-black">#{teamElo.rank}</p>
                  <p className="text-xs text-slate-500 mt-1">World Rank</p>
                </div>
                <div className="text-center">
                  <p className="text-sm font-semibold">{teamElo.date}</p>
                  <p className="text-xs text-slate-500 mt-1">As of Date</p>
                </div>
                <div className="text-center">
                  <p className="text-sm font-semibold">{teamElo.matchesConsidered}</p>
                  <p className="text-xs text-slate-500 mt-1">Matches</p>
                </div>
              </CardContent>
            </Card>
          )}
        </TabsContent>

        {/* TIMELINE */}
        <TabsContent value="timeline" className="space-y-4 mt-4">
          <Card>
            <CardContent className="p-4">
              <form onSubmit={handleTimeline} className="flex flex-wrap gap-3 items-end">
                <div className="space-y-1 flex-1 min-w-[160px]">
                  <Label>Team</Label>
                  <select
                    className="flex h-10 w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-950"
                    value={tlTeam}
                    onChange={(e) => setTlTeam(e.target.value)}
                    required
                  >
                    <option value="">Select team...</option>
                    {teams.map((t) => <option key={t.id} value={t.id}>{t.name}</option>)}
                  </select>
                </div>
                <div className="space-y-1">
                  <Label>From</Label>
                  <Input type="date" value={tlStart} onChange={(e) => setTlStart(e.target.value)} className="w-40" />
                </div>
                <div className="space-y-1">
                  <Label>To</Label>
                  <Input type="date" value={tlEnd} onChange={(e) => setTlEnd(e.target.value)} className="w-40" />
                </div>
                <div className="space-y-1">
                  <Label>Resolution</Label>
                  <select
                    className="flex h-10 rounded-md border border-slate-200 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-950"
                    value={tlResolution}
                    onChange={(e) => setTlResolution(e.target.value)}
                  >
                    <option value="match">Match</option>
                    <option value="month">Month</option>
                    <option value="year">Year</option>
                  </select>
                </div>
                <Button type="submit" disabled={tlLoading}>
                  {tlLoading ? "Loading..." : "Load Timeline"}
                </Button>
              </form>
            </CardContent>
          </Card>

          {tlError && <Alert variant="destructive"><AlertDescription>{tlError}</AlertDescription></Alert>}

          {timeline && (
            <div className="space-y-2">
              <p className="font-semibold">{timeline.teamName} ELO Timeline</p>
              <div className="overflow-x-auto rounded-lg border">
                <table className="w-full text-sm">
                  <thead className="bg-slate-50 border-b">
                    <tr>
                      <th className="text-left p-3 font-semibold">Date</th>
                      <th className="text-left p-3 font-semibold">Opponent</th>
                      <th className="text-center p-3 font-semibold">H/A</th>
                      <th className="text-center p-3 font-semibold">Result</th>
                      <th className="text-right p-3 font-semibold">ELO</th>
                      <th className="text-right p-3 font-semibold">Change</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(timeline.data || []).map((entry, i) => (
                      <tr key={i} className="border-b hover:bg-slate-50">
                        <td className="p-3">{entry.date}</td>
                        <td className="p-3">{entry.opponent}</td>
                        <td className="p-3 text-center">
                          <Badge variant={entry.homeAway === "H" ? "default" : entry.homeAway === "A" ? "secondary" : "outline"}>
                            {entry.homeAway}
                          </Badge>
                        </td>
                        <td className="p-3 text-center">
                          <Badge variant={entry.result === "W" ? "success" : entry.result === "L" ? "destructive" : "outline"}>
                            {entry.result}
                          </Badge>
                        </td>
                        <td className="p-3 text-right tabular-nums font-semibold">{Math.round(entry.elo)}</td>
                        <td className="p-3 text-right tabular-nums">
                          <span className={entry.change > 0 ? "text-green-600" : entry.change < 0 ? "text-red-500" : "text-slate-500"}>
                            {entry.change > 0 ? "+" : ""}{Math.round(entry.change)}
                          </span>
                        </td>
                      </tr>
                    ))}
                    {(!timeline.data || timeline.data.length === 0) && (
                      <tr><td colSpan={6} className="p-8 text-center text-slate-500">No timeline data.</td></tr>
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </TabsContent>
      </Tabs>
    </div>
  )
}
