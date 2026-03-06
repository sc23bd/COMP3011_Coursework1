// Package elo provides the World Football Elo rating calculation engine.
//
// Formula:
//
//	Elo_new = Elo_old + K * GoalMarginAdj * (ActualResult - ExpectedResult)
//
// where:
//
//	ExpectedResult   = 1 / (10^(-dr/400) + 1),  dr = homeElo - awayElo + homeAdvantage
//	GoalMarginAdj    = piecewise multiplier on goal difference N (see GoalMarginMultiplier)
//	ActualResult     = 1.0 (win), 0.5 (draw), 0.0 (loss)   from home-team perspective
//	K                = tournament-dependent factor (see Config.KFactor)
//
// Reference: https://www.eloratings.net/about
package elo

import (
	"math"
	"sort"
	"time"
)

// RatingMap maps team IDs to their current Elo rating.
type RatingMap map[int]float64

// Calculate replays matches chronologically and returns the Elo rating for
// every team that participated in at least one match up to (and including) endDate.
//
// matches must not be nil; pass an empty slice to get an empty result.
// Teams with no prior matches start at cfg.DefaultRating.
func Calculate(matches []MatchResult, cfg Config) RatingMap {
	// Sort ascending by date then by match ID for deterministic ordering.
	sorted := make([]MatchResult, len(matches))
	copy(sorted, matches)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Date.Equal(sorted[j].Date) {
			return sorted[i].MatchID < sorted[j].MatchID
		}
		return sorted[i].Date.Before(sorted[j].Date)
	})

	ratings := make(RatingMap)
	for _, m := range sorted {
		processMatch(m, ratings, cfg)
	}
	return ratings
}

// CalculateUntil is like Calculate but only considers matches up to and
// including the given date.
func CalculateUntil(matches []MatchResult, endDate time.Time, cfg Config) RatingMap {
	filtered := make([]MatchResult, 0, len(matches))
	for _, m := range matches {
		if !m.Date.After(endDate) {
			filtered = append(filtered, m)
		}
	}
	return Calculate(filtered, cfg)
}

// CalculateTimeline replays matches for a single team and returns a
// TimelineEntry for every match involving that team.
func CalculateTimeline(teamID int, matches []MatchResult, cfg Config) []TimelineEntry {
	sorted := make([]MatchResult, len(matches))
	copy(sorted, matches)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Date.Equal(sorted[j].Date) {
			return sorted[i].MatchID < sorted[j].MatchID
		}
		return sorted[i].Date.Before(sorted[j].Date)
	})

	ratings := make(RatingMap)
	var timeline []TimelineEntry

	for _, m := range sorted {
		prevHome := rating(ratings, m.HomeTeamID, cfg.DefaultRating)
		prevAway := rating(ratings, m.AwayTeamID, cfg.DefaultRating)

		processMatch(m, ratings, cfg)

		if m.HomeTeamID == teamID {
			entry := buildEntry(m, ratings[teamID], prevHome, m.AwayTeamID, true)
			timeline = append(timeline, entry)
		} else if m.AwayTeamID == teamID {
			entry := buildEntry(m, ratings[teamID], prevAway, m.HomeTeamID, false)
			timeline = append(timeline, entry)
		}
	}
	return timeline
}

// ExpectedResult calculates the expected outcome (0–1) for the home team.
//
//	dr  = homeElo - awayElo + homeAdvantage
//	E   = 1 / (10^(-dr/400) + 1)
func ExpectedResult(homeElo, awayElo, homeAdvantage float64) float64 {
	dr := homeElo - awayElo + homeAdvantage
	return 1.0 / (math.Pow(10, -dr/400) + 1)
}

// GoalMarginMultiplier returns the K-factor scaling multiplier for the absolute
// goal difference N, following the exact World Football Elo step formula:
//
//   - N ≤ 1  → 1.0   (no adjustment)
//   - N == 2 → 1.5   (K increased by half)
//   - N == 3 → 1.75  (K increased by 3/4)
//   - N ≥ 4  → 1.75 + (N-3)/8
func GoalMarginMultiplier(goalDiff int) float64 {
	n := goalDiff
	if n < 0 {
		n = -n
	}
	switch {
	case n <= 1:
		return 1.0
	case n == 2:
		return 1.5
	case n == 3:
		return 1.75
	default:
		return 1.75 + float64(n-3)/8.0
	}
}

// --- internal helpers -------------------------------------------------------

// processMatch applies one match result to the ratings map in place.
func processMatch(m MatchResult, ratings RatingMap, cfg Config) {
	homeElo := rating(ratings, m.HomeTeamID, cfg.DefaultRating)
	awayElo := rating(ratings, m.AwayTeamID, cfg.DefaultRating)

	advantage := cfg.HomeAdvantage
	if m.Neutral {
		advantage = 0
	}

	expected := ExpectedResult(homeElo, awayElo, advantage)
	k := cfg.KFactor(m.Tournament)
	gmAdj := GoalMarginMultiplier(m.HomeScore - m.AwayScore)

	var actual float64
	switch {
	case m.HomeScore > m.AwayScore:
		actual = 1.0
	case m.HomeScore < m.AwayScore:
		actual = 0.0
	default:
		actual = 0.5
	}

	delta := k * gmAdj * (actual - expected)
	ratings[m.HomeTeamID] = homeElo + delta
	ratings[m.AwayTeamID] = awayElo - delta
}

// rating returns the current rating for a team, defaulting when not yet seen.
func rating(ratings RatingMap, teamID int, def float64) float64 {
	if r, ok := ratings[teamID]; ok {
		return r
	}
	return def
}

// buildEntry creates a TimelineEntry from a single match involving the focus team.
func buildEntry(m MatchResult, newElo, prevElo float64, opponentID int, isHome bool) TimelineEntry {
	_ = opponentID // used for caller context; opponent name set by caller when available
	var homeAway string
	if m.Neutral {
		homeAway = "N"
	} else if isHome {
		homeAway = "H"
	} else {
		homeAway = "A"
	}

	// Determine result from focus-team perspective.
	var result string
	switch {
	case m.HomeScore == m.AwayScore:
		result = "D"
	case isHome && m.HomeScore > m.AwayScore:
		result = "W"
	case !isHome && m.AwayScore > m.HomeScore:
		result = "W"
	default:
		result = "L"
	}

	return TimelineEntry{
		Date:     m.Date,
		Elo:      math.Round(newElo*100) / 100,
		Change:   math.Round((newElo-prevElo)*100) / 100,
		MatchID:  m.MatchID,
		Result:   result,
		HomeAway: homeAway,
	}
}
