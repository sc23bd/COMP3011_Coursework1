import React, { useState, useEffect, useCallback } from "react"
import { useParams, useNavigate } from "react-router-dom"
import { useAuth } from "@/hooks/useAuth"
import {
  getMatch, getGoals, createGoal, deleteGoal,
  getShootout, createShootout, deleteShootout, getTeams
} from "@/api/client"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogClose
} from "@/components/ui/dialog"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { ArrowLeft, Plus, Trash2, Target, Shield } from "lucide-react"

function AddGoalDialog({ open, onClose, onSave, matchId, homeTeamId, homeName, awayTeamId, awayName }) {
  const [form, setForm] = useState({ scorer: "", teamId: "", ownGoal: false, penalty: false })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState("")

  useEffect(() => { setForm({ scorer: "", teamId: "", ownGoal: false, penalty: false }); setError("") }, [open])

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError("")
    try {
      await onSave({ ...form, teamId: parseInt(form.teamId) })
      onClose()
    } catch (err) {
      setError(err.response?.data?.error || "Failed to add goal")
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader><DialogTitle>Add Goal</DialogTitle></DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>Scorer Name</Label>
            <Input value={form.scorer} onChange={(e) => setForm({ ...form, scorer: e.target.value })} placeholder="e.g. Messi" required minLength={1} maxLength={100} />
          </div>
          <div className="space-y-2">
            <Label>Team</Label>
            <select
              className="flex h-10 w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-950"
              value={form.teamId}
              onChange={(e) => setForm({ ...form, teamId: e.target.value })}
              required
            >
              <option value="">Select team...</option>
              {homeTeamId && <option value={homeTeamId}>{homeName}</option>}
              {awayTeamId && <option value={awayTeamId}>{awayName}</option>}
            </select>
          </div>
          <div className="flex gap-4">
            <label className="flex items-center gap-2 text-sm cursor-pointer">
              <input type="checkbox" checked={form.ownGoal} onChange={(e) => setForm({ ...form, ownGoal: e.target.checked })} />
              Own Goal
            </label>
            <label className="flex items-center gap-2 text-sm cursor-pointer">
              <input type="checkbox" checked={form.penalty} onChange={(e) => setForm({ ...form, penalty: e.target.checked })} />
              Penalty
            </label>
          </div>
          {error && <Alert variant="destructive"><AlertDescription>{error}</AlertDescription></Alert>}
          <DialogFooter>
            <DialogClose asChild><Button type="button" variant="outline">Cancel</Button></DialogClose>
            <Button type="submit" disabled={loading}>{loading ? "Adding..." : "Add Goal"}</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

function AddShootoutDialog({ open, onClose, onSave, homeTeamId, homeName, awayTeamId, awayName }) {
  const [winnerId, setWinnerId] = useState("")
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState("")

  useEffect(() => { setWinnerId(""); setError("") }, [open])

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError("")
    try {
      await onSave({ winnerId: parseInt(winnerId) })
      onClose()
    } catch (err) {
      setError(err.response?.data?.error || "Failed to add shootout")
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader><DialogTitle>Record Penalty Shootout</DialogTitle></DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>Winning Team</Label>
            <select
              className="flex h-10 w-full rounded-md border border-slate-200 bg-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-slate-950"
              value={winnerId}
              onChange={(e) => setWinnerId(e.target.value)}
              required
            >
              <option value="">Select winner...</option>
              {homeTeamId && <option value={homeTeamId}>{homeName}</option>}
              {awayTeamId && <option value={awayTeamId}>{awayName}</option>}
            </select>
          </div>
          {error && <Alert variant="destructive"><AlertDescription>{error}</AlertDescription></Alert>}
          <DialogFooter>
            <DialogClose asChild><Button type="button" variant="outline">Cancel</Button></DialogClose>
            <Button type="submit" disabled={loading}>{loading ? "Saving..." : "Save"}</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

export default function MatchDetailPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { isAuthenticated } = useAuth()
  const [match, setMatch] = useState(null)
  const [goals, setGoals] = useState([])
  const [shootout, setShootout] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [addGoalOpen, setAddGoalOpen] = useState(false)
  const [addShootoutOpen, setAddShootoutOpen] = useState(false)

  const loadData = useCallback(async () => {
    setLoading(true)
    setError("")
    try {
      const [matchRes, goalsRes] = await Promise.all([getMatch(id), getGoals(id)])
      setMatch(matchRes.data)
      setGoals(goalsRes.data?.data || [])
      try {
        const soRes = await getShootout(id)
        setShootout(soRes.data)
      } catch {
        setShootout(null)
      }
    } catch (err) {
      setError(err.response?.data?.error || "Failed to load match")
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => { loadData() }, [loadData])

  const handleAddGoal = async (data) => {
    await createGoal(id, data)
    await loadData()
  }

  const handleDeleteGoal = async (goalId) => {
    if (!window.confirm("Delete this goal?")) return
    try {
      await deleteGoal(id, goalId)
      await loadData()
    } catch (err) {
      setError(err.response?.data?.error || "Delete failed")
    }
  }

  const handleAddShootout = async (data) => {
    await createShootout(id, data)
    await loadData()
  }

  const handleDeleteShootout = async () => {
    if (!window.confirm("Delete shootout record?")) return
    try {
      await deleteShootout(id)
      await loadData()
    } catch (err) {
      setError(err.response?.data?.error || "Delete failed")
    }
  }

  if (loading) return <p className="text-slate-500 text-center py-16">Loading match...</p>
  if (error) return <Alert variant="destructive"><AlertDescription>{error}</AlertDescription></Alert>
  if (!match) return null

  const homeGoals = goals.filter((g) => g.teamId === match.homeTeamId && !g.ownGoal)
    .concat(goals.filter((g) => g.teamId === match.awayTeamId && g.ownGoal))
  const awayGoals = goals.filter((g) => g.teamId === match.awayTeamId && !g.ownGoal)
    .concat(goals.filter((g) => g.teamId === match.homeTeamId && g.ownGoal))

  return (
    <div className="space-y-6 max-w-3xl mx-auto">
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="icon" onClick={() => navigate(-1)}>
          <ArrowLeft className="h-5 w-5" />
        </Button>
        <h2 className="text-2xl font-bold">Match Details</h2>
      </div>

      {/* Match header */}
      <Card>
        <CardContent className="p-6">
          <div className="flex justify-between items-center mb-4">
            <Badge variant="secondary">{match.tournament}</Badge>
            <span className="text-sm text-slate-500">{match.date} · {match.city}, {match.country}</span>
          </div>
          <div className="grid grid-cols-3 items-center gap-4 text-center">
            <div>
              <p className="text-lg font-bold">{match.homeTeam}</p>
              <p className="text-xs text-slate-500">Home</p>
            </div>
            <div className="text-4xl font-black tabular-nums">
              {match.homeScore} <span className="text-slate-400">–</span> {match.awayScore}
            </div>
            <div>
              <p className="text-lg font-bold">{match.awayTeam}</p>
              <p className="text-xs text-slate-500">Away</p>
            </div>
          </div>
          {match.neutral && <Badge variant="outline" className="mt-3 mx-auto block w-fit">Neutral venue</Badge>}
        </CardContent>
      </Card>

      <Tabs defaultValue="goals">
        <TabsList>
          <TabsTrigger value="goals">Goals ({goals.length})</TabsTrigger>
          <TabsTrigger value="shootout">Shootout</TabsTrigger>
        </TabsList>

        <TabsContent value="goals" className="space-y-3 mt-4">
          {isAuthenticated && (
            <Button size="sm" onClick={() => setAddGoalOpen(true)}>
              <Plus className="h-4 w-4 mr-1" /> Add Goal
            </Button>
          )}
          {goals.length === 0 ? (
            <p className="text-slate-500 text-center py-8">No goals recorded.</p>
          ) : (
            <div className="space-y-2">
              {goals.map((g) => (
                <Card key={g.id}>
                  <CardContent className="p-3 flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <Target className="h-4 w-4 text-slate-400 shrink-0" />
                      <div>
                        <span className="font-medium">{g.scorer}</span>
                        <span className="text-slate-500 text-sm"> · {g.team}</span>
                        <div className="flex gap-1 mt-0.5">
                          {g.ownGoal && <Badge variant="destructive" className="text-xs">OG</Badge>}
                          {g.penalty && <Badge variant="secondary" className="text-xs">Pen</Badge>}
                        </div>
                      </div>
                    </div>
                    {isAuthenticated && (
                      <Button size="icon" variant="ghost" className="text-red-500 hover:text-red-700" onClick={() => handleDeleteGoal(g.id)}>
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    )}
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </TabsContent>

        <TabsContent value="shootout" className="mt-4">
          {shootout ? (
            <Card>
              <CardContent className="p-4 space-y-2">
                <div className="flex items-center gap-2">
                  <Shield className="h-5 w-5 text-yellow-500" />
                  <span className="font-semibold">Penalty Shootout Winner: {shootout.winner}</span>
                </div>
                {isAuthenticated && (
                  <Button size="sm" variant="destructive" onClick={handleDeleteShootout}>
                    <Trash2 className="h-4 w-4 mr-1" /> Delete
                  </Button>
                )}
              </CardContent>
            </Card>
          ) : (
            <div className="text-center py-8 space-y-3">
              <p className="text-slate-500">No shootout recorded for this match.</p>
              {isAuthenticated && (
                <Button size="sm" onClick={() => setAddShootoutOpen(true)}>
                  <Plus className="h-4 w-4 mr-1" /> Record Shootout
                </Button>
              )}
            </div>
          )}
        </TabsContent>
      </Tabs>

      <AddGoalDialog
        open={addGoalOpen}
        onClose={() => setAddGoalOpen(false)}
        onSave={handleAddGoal}
        matchId={id}
        homeTeamId={match.homeTeamId}
        homeName={match.homeTeam}
        awayTeamId={match.awayTeamId}
        awayName={match.awayTeam}
      />
      <AddShootoutDialog
        open={addShootoutOpen}
        onClose={() => setAddShootoutOpen(false)}
        onSave={handleAddShootout}
        homeTeamId={match.homeTeamId}
        homeName={match.homeTeam}
        awayTeamId={match.awayTeamId}
        awayName={match.awayTeam}
      />
    </div>
  )
}
