import React, { useState } from "react"
import { getPlayerGoals } from "@/api/client"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Search, Target } from "lucide-react"

export default function PlayersPage() {
  const [name, setName] = useState("")
  const [goals, setGoals] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState("")

  const handleSearch = async (e) => {
    e.preventDefault()
    if (!name.trim()) return
    setLoading(true)
    setError("")
    setGoals(null)
    try {
      const res = await getPlayerGoals(name.trim())
      setGoals(res.data?.data || [])
    } catch (err) {
      setError(err.response?.data?.error || "Player not found")
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6 max-w-2xl">
      <h2 className="text-2xl font-bold">Player Goals</h2>

      <Card>
        <CardContent className="p-4">
          <form onSubmit={handleSearch} className="flex gap-2">
            <div className="flex-1 space-y-1">
              <Label>Player Name</Label>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. Messi"
                required
              />
            </div>
            <div className="pt-6">
              <Button type="submit" disabled={loading}>
                <Search className="h-4 w-4 mr-1" />
                {loading ? "Searching..." : "Search"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>

      {error && <Alert variant="destructive"><AlertDescription>{error}</AlertDescription></Alert>}

      {goals !== null && (
        <div className="space-y-2">
          <p className="text-sm text-slate-500">
            {goals.length === 0 ? "No goals found" : `${goals.length} goal(s) found`} for <strong>{name}</strong>
          </p>
          {goals.map((g) => (
            <Card key={g.id}>
              <CardContent className="p-3 flex items-center gap-3">
                <Target className="h-4 w-4 text-slate-400 shrink-0" />
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <span className="font-medium">{g.scorer}</span>
                    <span className="text-slate-500 text-sm">·</span>
                    <span className="text-slate-500 text-sm">{g.team}</span>
                    {g.ownGoal && <Badge variant="destructive" className="text-xs">OG</Badge>}
                    {g.penalty && <Badge variant="secondary" className="text-xs">Pen</Badge>}
                  </div>
                  <p className="text-xs text-slate-400">Match ID: {g.matchId}</p>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
