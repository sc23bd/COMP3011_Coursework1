import React, { useState, useEffect, useCallback } from "react"
import { getTeams, getTeam, createTeam, updateTeam, deleteTeam, getTeamHistory } from "@/api/client"
import { useAuth } from "@/hooks/useAuth"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Alert, AlertDescription } from "@/components/ui/alert"
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogClose
} from "@/components/ui/dialog"
import { Badge } from "@/components/ui/badge"
import { Pencil, Trash2, Plus, Eye, History } from "lucide-react"
import { formatDate } from "@/lib/utils"

function TeamDialog({ open, onClose, onSave, team }) {
  const [name, setName] = useState(team?.name || "")
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState("")

  useEffect(() => { setName(team?.name || ""); setError("") }, [team, open])

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    setError("")
    try {
      await onSave(name)
      onClose()
    } catch (err) {
      setError(err.response?.data?.error || "Failed to save team")
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{team ? "Edit Team" : "Create Team"}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>Team Name</Label>
            <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g. England" required minLength={1} maxLength={100} />
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

function TeamDetailDialog({ open, onClose, teamId }) {
  const [detail, setDetail] = useState(null)
  const [history, setHistory] = useState(null)

  useEffect(() => {
    if (!open || !teamId) return
    let isMounted = true
    Promise.all([getTeam(teamId), getTeamHistory(teamId)])
      .then(([teamRes, histRes]) => {
        if (isMounted) {
          setDetail(teamRes.data)
          setHistory(histRes.data)
        }
      })
      .catch(() => {})
    return () => {
      isMounted = false
      setDetail(null)
      setHistory(null)
    }
  }, [open, teamId])

  const loading = open && !!teamId && !detail

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Team Details</DialogTitle>
        </DialogHeader>
        {loading ? (
          <p className="text-center py-4 text-slate-500">Loading...</p>
        ) : detail ? (
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-2 text-sm">
              <span className="font-medium text-slate-500">ID</span><span>{detail.id}</span>
              <span className="font-medium text-slate-500">Name</span><span className="font-semibold">{detail.name}</span>
              <span className="font-medium text-slate-500">Created</span>
              <span>{detail.createdAt ? formatDate(detail.createdAt) : "—"}</span>
            </div>
            {history?.data?.length > 0 && (
              <div>
                <p className="font-semibold text-sm mb-2">Former Names</p>
                <div className="space-y-1">
                  {history.data.map((h) => (
                    <div key={h.id} className="flex justify-between text-sm border rounded p-2">
                      <span>{h.formerName}</span>
                      <span className="text-slate-500">{formatDate(h.startDate)} — {h.endDate ? formatDate(h.endDate) : "present"}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        ) : null}
      </DialogContent>
    </Dialog>
  )
}

export default function TeamsPage() {
  const { isAuthenticated } = useAuth()
  const [teams, setTeams] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [search, setSearch] = useState("")
  const [createOpen, setCreateOpen] = useState(false)
  const [editTeam, setEditTeam] = useState(null)
  const [viewTeamId, setViewTeamId] = useState(null)

  const loadTeams = useCallback(async () => {
    setLoading(true)
    setError("")
    try {
      const res = await getTeams()
      setTeams(res.data?.data || [])
    } catch {
      setError("Failed to load teams")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { loadTeams() }, [loadTeams])

  const handleCreate = async (name) => {
    await createTeam({ name })
    await loadTeams()
  }

  const handleEdit = async (name) => {
    await updateTeam(editTeam.id, { name })
    await loadTeams()
  }

  const handleDelete = async (team) => {
    if (!window.confirm(`Delete "${team.name}"?`)) return
    try {
      await deleteTeam(team.id)
      await loadTeams()
    } catch (err) {
      setError(err.response?.data?.error || "Delete failed")
    }
  }

  const filtered = teams.filter((t) =>
    t.name.toLowerCase().includes(search.toLowerCase())
  )

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold">Teams</h2>
        {isAuthenticated && (
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <Plus className="h-4 w-4 mr-1" /> New Team
          </Button>
        )}
      </div>

      <Input
        placeholder="Search teams..."
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        className="max-w-sm"
      />

      {error && <Alert variant="destructive"><AlertDescription>{error}</AlertDescription></Alert>}

      {loading ? (
        <p className="text-slate-500 text-center py-8">Loading teams...</p>
      ) : (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {filtered.map((team) => (
            <Card key={team.id} className="group">
              <CardContent className="p-4 flex items-center justify-between">
                <div>
                  <p className="font-semibold">{team.name}</p>
                  <p className="text-xs text-slate-500">ID: {team.id}</p>
                </div>
                <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  <Button size="icon" variant="ghost" onClick={() => setViewTeamId(team.id)} title="View">
                    <Eye className="h-4 w-4" />
                  </Button>
                  {isAuthenticated && (
                    <>
                      <Button size="icon" variant="ghost" onClick={() => setEditTeam(team)} title="Edit">
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button size="icon" variant="ghost" className="text-red-500 hover:text-red-700" onClick={() => handleDelete(team)} title="Delete">
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </>
                  )}
                </div>
              </CardContent>
            </Card>
          ))}
          {filtered.length === 0 && !loading && (
            <p className="text-slate-500 col-span-full text-center py-8">No teams found.</p>
          )}
        </div>
      )}

      <TeamDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onSave={handleCreate}
        team={null}
      />
      <TeamDialog
        open={!!editTeam}
        onClose={() => setEditTeam(null)}
        onSave={handleEdit}
        team={editTeam}
      />
      <TeamDetailDialog
        open={!!viewTeamId}
        onClose={() => setViewTeamId(null)}
        teamId={viewTeamId}
      />
    </div>
  )
}
