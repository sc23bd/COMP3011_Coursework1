import React, { useState, useEffect, useCallback } from "react"
import { useNavigate } from "react-router-dom"
import { getMatches, getHeadToHead, getTeams } from "@/api/client"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Eye, ChevronLeft, ChevronRight } from "lucide-react"

function MatchCard({ match, onClick }) {
  return (
    <Card className="cursor-pointer hover:shadow-md transition-shadow" onClick={onClick}>
      <CardContent className="p-4">
        <div className="flex justify-between items-center mb-2">
          <Badge variant="secondary">{match.tournament}</Badge>
          <span className="text-xs text-slate-500">{match.date}</span>
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

export default function MatchesPage() {
  const navigate = useNavigate()
  const [matches, setMatches] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [limit] = useState(50)
  const [offset, setOffset] = useState(0)

  // Head to head
  const [teams, setTeams] = useState([])
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

  return (
    <div className="space-y-4">
      <h2 className="text-2xl font-bold">Matches</h2>

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
    </div>
  )
}
