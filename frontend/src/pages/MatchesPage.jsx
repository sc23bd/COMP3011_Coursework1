import React, { useState, useEffect, useCallback } from "react"
import { useNavigate } from "react-router-dom"
import { getMatches, getHeadToHead, getTeams, createMatch, getTournaments } from "@/api/client"
import { useAuth } from "@/hooks/useAuth"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogClose
} from "@/components/ui/dialog"
import { ChevronLeft, ChevronRight, Plus } from "lucide-react"
import { formatDate } from "@/lib/utils"

function MatchCard({ match, onClick }) {
  return (
    <Card className="cursor-pointer hover:shadow-md transition-shadow" onClick={onClick}>
      <CardContent className="p-4">
        <div className="flex justify-between items-center mb-2">
          <Badge variant="secondary">{match.tournament}</Badge>
          <span className="text-xs text-slate-500">{formatDate(match.date)}</span>
        </div>
        <div className="flex items-center justify-between gap-2">
          <span className="font-semibold text-right flex-1 truncate">{match.homeTeam}</span>
          <div className="flex items-center gap-1 shrink-0">
            <span className="text-xl font-bold tabular-nums">{match.homeScore}</span>
            <span className="text-slate-400">–</span>
            <span className="text-xl font-bold tabular-nums">{match.awayScore}</span>
          </div>
          <span className="font-semibold text-left flex-1 truncate">{match.awayTeam}</span>
        </div>
        <p className="text-xs text-slate-500 mt-1 text-center">{match.city}, {match.country}</p>
      </CardContent>
    </Card>
  )
}

const EMPTY_MATCH_FORM = { date: "", homeTeamId: "", awayTeamId: "", homeScore: 0, awayScore: 0, tournamentId: "", city: "", country: "", neutral: false }

function CreateMatchDialog({ open, onClose, onSave, teams, tournaments }) {
  const [form, setForm] = useState(EMPTY_MATCH_FORM)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState("")

  useEffect(() => { setForm(EMPTY_MATCH_FORM); setError("") }, [open])

  const set = (key, val) => setForm((f) => ({ ...f, [key]: val }))

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError("")
    try {
      // Use the date string directly with T00:00:00Z to avoid timezone shifts
      // that occur when converting a date-only string via new Date().toISOString().
      await onSave({
        date: `${form.date}T00:00:00Z`,
        homeTeamId: parseInt(form.homeTeamId),
        awayTeamId: parseInt(form.awayTeamId),
        homeScore: parseInt(form.homeScore),
        awayScore: parseInt(form.awayScore),
        tournamentId: parseInt(form.tournamentId),
        city: form.city,
        country: form.country,
        neutral: form.neutral,
      })
      onClose()
    } catch (err) {
      setError(err.response?.data?.error || "Failed to create match")
    } finally {
      setLoading(false)
    }
  }

  const selectClass = "flex h-10 w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-950"

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-lg max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>New Match</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>Date</Label>
            <Input type="date" value={form.date} onChange={(e) => set("date", e.target.value)} required />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-2">
              <Label>Home Team</Label>
              <select className={selectClass} value={form.homeTeamId} onChange={(e) => set("homeTeamId", e.target.value)} required>
                <option value="">Select...</option>
                {teams.map((t) => <option key={t.id} value={t.id}>{t.name}</option>)}
              </select>
            </div>
            <div className="space-y-2">
              <Label>Away Team</Label>
              <select className={selectClass} value={form.awayTeamId} onChange={(e) => set("awayTeamId", e.target.value)} required>
                <option value="">Select...</option>
                {teams.map((t) => <option key={t.id} value={t.id}>{t.name}</option>)}
              </select>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-2">
              <Label>Home Score</Label>
              <Input type="number" min="0" value={form.homeScore} onChange={(e) => set("homeScore", e.target.value)} required />
            </div>
            <div className="space-y-2">
              <Label>Away Score</Label>
              <Input type="number" min="0" value={form.awayScore} onChange={(e) => set("awayScore", e.target.value)} required />
            </div>
          </div>
          <div className="space-y-2">
            <Label>Tournament</Label>
            <select className={selectClass} value={form.tournamentId} onChange={(e) => set("tournamentId", e.target.value)} required>
              <option value="">Select...</option>
              {tournaments.map((t) => <option key={t.id} value={t.id}>{t.name}</option>)}
            </select>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-2">
              <Label>City</Label>
              <Input value={form.city} onChange={(e) => set("city", e.target.value)} placeholder="e.g. London" />
            </div>
            <div className="space-y-2">
              <Label>Country</Label>
              <Input value={form.country} onChange={(e) => set("country", e.target.value)} placeholder="e.g. England" />
            </div>
          </div>
          <label className="flex items-center gap-2 text-sm cursor-pointer">
            <input type="checkbox" checked={form.neutral} onChange={(e) => set("neutral", e.target.checked)} />
            Neutral venue
          </label>
          {error && <Alert variant="destructive"><AlertDescription>{error}</AlertDescription></Alert>}
          <DialogFooter>
            <DialogClose asChild><Button type="button" variant="outline">Cancel</Button></DialogClose>
            <Button type="submit" disabled={loading}>{loading ? "Creating..." : "Create Match"}</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

export default function MatchesPage() {
  const navigate = useNavigate()
  const { isAuthenticated } = useAuth()
  const [matches, setMatches] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [limit] = useState(50)
  const [offset, setOffset] = useState(0)
  const [createOpen, setCreateOpen] = useState(false)

  // Head to head
  const [teams, setTeams] = useState([])
  const [tournaments, setTournaments] = useState([])
  const [teamAId, setTeamAId] = useState("")
  const [teamBId, setTeamBId] = useState("")
  const [h2hMatches, setH2hMatches] = useState(null)
  const [h2hLoading, setH2hLoading] = useState(false)
  const [h2hError, setH2hError] = useState("")

  const loadMatches = useCallback(async () => {
    setLoading(true)
    setError("")
    try {
      const res = await getMatches({ limit, offset })
      setMatches(res.data?.data || [])
    } catch {
      setError("Failed to load matches")
    } finally {
      setLoading(false)
    }
  }, [limit, offset])

  useEffect(() => { loadMatches() }, [loadMatches])

  useEffect(() => {
    getTeams().then((r) => setTeams(r.data?.data || [])).catch(() => {})
    getTournaments().then((r) => setTournaments(r.data?.data || [])).catch(() => {})
  }, [])

  const handleH2H = async (e) => {
    e.preventDefault()
    if (!teamAId || !teamBId) return
    setH2hLoading(true)
    setH2hError("")
    setH2hMatches(null)
    try {
      const res = await getHeadToHead(teamAId, teamBId)
      setH2hMatches(res.data?.data || [])
    } catch (err) {
      setH2hError(err.response?.data?.error || "Failed to load head-to-head data")
    } finally {
      setH2hLoading(false)
    }
  }

  const handleCreateMatch = async (data) => {
    await createMatch(data)
    await loadMatches()
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold">Matches</h2>
        {isAuthenticated && (
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <Plus className="h-4 w-4 mr-1" /> New Match
          </Button>
        )}
      </div>

      <Tabs defaultValue="list">
        <TabsList>
          <TabsTrigger value="list">All Matches</TabsTrigger>
          <TabsTrigger value="h2h">Head to Head</TabsTrigger>
        </TabsList>

        <TabsContent value="list" className="space-y-4 mt-4">
          {error && <Alert variant="destructive"><AlertDescription>{error}</AlertDescription></Alert>}
          {loading ? (
            <p className="text-slate-500 text-center py-8">Loading matches...</p>
          ) : (
            <>
              <div className="grid gap-3 sm:grid-cols-2">
                {matches.map((m) => (
                  <MatchCard key={m.id} match={m} onClick={() => navigate(`/matches/${m.id}`)} />
                ))}
                {matches.length === 0 && (
                  <p className="text-slate-500 col-span-full text-center py-8">No matches found.</p>
                )}
              </div>
              <div className="flex items-center justify-between pt-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setOffset(Math.max(0, offset - limit))}
                  disabled={offset === 0}
                >
                  <ChevronLeft className="h-4 w-4 mr-1" /> Previous
                </Button>
                <span className="text-sm text-slate-500">Showing {offset + 1}–{offset + matches.length}</span>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setOffset(offset + limit)}
                  disabled={matches.length < limit}
                >
                  Next <ChevronRight className="h-4 w-4 ml-1" />
                </Button>
              </div>
            </>
          )}
        </TabsContent>

        <TabsContent value="h2h" className="space-y-4 mt-4">
          <Card>
            <CardHeader><CardTitle className="text-base">Compare Two Teams</CardTitle></CardHeader>
            <CardContent>
              <form onSubmit={handleH2H} className="flex flex-wrap gap-3 items-end">
                <div className="space-y-1 flex-1 min-w-[160px]">
                  <Label>Team A</Label>
                  <select
                    className="flex h-10 w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-950"
                    value={teamAId}
                    onChange={(e) => setTeamAId(e.target.value)}
                    required
                  >
                    <option value="">Select team...</option>
                    {teams.map((t) => <option key={t.id} value={t.id}>{t.name}</option>)}
                  </select>
                </div>
                <div className="space-y-1 flex-1 min-w-[160px]">
                  <Label>Team B</Label>
                  <select
                    className="flex h-10 w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-950"
                    value={teamBId}
                    onChange={(e) => setTeamBId(e.target.value)}
                    required
                  >
                    <option value="">Select team...</option>
                    {teams.map((t) => <option key={t.id} value={t.id}>{t.name}</option>)}
                  </select>
                </div>
                <Button type="submit" disabled={h2hLoading}>
                  {h2hLoading ? "Loading..." : "Compare"}
                </Button>
              </form>
            </CardContent>
          </Card>

          {h2hError && <Alert variant="destructive"><AlertDescription>{h2hError}</AlertDescription></Alert>}

          {h2hMatches && (
            <div>
              <p className="text-sm text-slate-500 mb-3">{h2hMatches.length} match(es) found</p>
              <div className="grid gap-3 sm:grid-cols-2">
                {h2hMatches.map((m) => (
                  <MatchCard key={m.id} match={m} onClick={() => navigate(`/matches/${m.id}`)} />
                ))}
                {h2hMatches.length === 0 && (
                  <p className="text-slate-500 col-span-full text-center py-8">No matches between these teams.</p>
                )}
              </div>
            </div>
          )}
        </TabsContent>
      </Tabs>

      <CreateMatchDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onSave={handleCreateMatch}
        teams={teams}
        tournaments={tournaments}
      />
    </div>
  )
}
