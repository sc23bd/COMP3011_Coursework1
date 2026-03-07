import React from "react"
import { useNavigate } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Users, Calendar, User, TrendingUp, ArrowRight } from "lucide-react"

const sections = [
  {
    to: "/teams",
    Icon: Users,
    title: "Teams",
    description: "Browse, create, edit, and manage national football teams.",
    color: "text-blue-600 bg-blue-50",
  },
  {
    to: "/matches",
    Icon: Calendar,
    title: "Matches",
    description: "View match results, head-to-head comparisons, goals and shootouts.",
    color: "text-green-600 bg-green-50",
  },
  {
    to: "/players",
    Icon: User,
    title: "Players",
    description: "Search goals scored by any player across all matches.",
    color: "text-purple-600 bg-purple-50",
  },
  {
    to: "/elo",
    Icon: TrendingUp,
    title: "ELO Rankings",
    description: "View global Elo ratings, team rankings, and historical timelines.",
    color: "text-orange-600 bg-orange-50",
  },
]

export default function HomePage() {
  const navigate = useNavigate()
  return (
    <div className="space-y-8">
      <div className="text-center space-y-2">
        <h1 className="text-4xl font-black text-slate-900">⚽ Football API Explorer</h1>
        <p className="text-slate-500 max-w-xl mx-auto">
          A modern interface for exploring football match data, team statistics, player goals, and ELO rankings.
        </p>
      </div>
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {sections.map(({ to, Icon, title, description, color }) => (
          <Card
            key={to}
            className="cursor-pointer hover:shadow-md transition-shadow"
            onClick={() => navigate(to)}
          >
            <CardHeader className="pb-2">
              <div className={`w-10 h-10 rounded-lg flex items-center justify-center mb-2 ${color}`}>
                <Icon className="h-5 w-5" />
              </div>
              <CardTitle className="text-base">{title}</CardTitle>
            </CardHeader>
            <CardContent>
              <CardDescription>{description}</CardDescription>
              <Button variant="link" className="p-0 mt-2 h-auto text-slate-900" onClick={() => navigate(to)}>
                Explore <ArrowRight className="h-3.5 w-3.5 ml-1" />
              </Button>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}
