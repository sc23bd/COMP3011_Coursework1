// Package simulator implements a Monte Carlo match-outcome simulation engine
// for international football.
//
// # Methodology
//
// Each simulation iteration draws independent Poisson-distributed goal totals
// for the home and away teams.  The Poisson rate parameters (λ) are derived
// from two sources:
//
//  1. Historical scoring averages – the mean number of goals a team has scored
//     per game over their recorded match history.  When no history is available
//     the engine falls back to BaseGoalRate (1.25 goals/game, the approximate
//     average in international football).
//
//  2. Elo rating differential – the Elo formula's expected result E (a value
//     in [0,1]) captures the relative strength of the two teams.  E is used to
//     shift λ away from the historical average in proportion to the mismatch:
//
//     λ_home = homeGoalRate × (1 + EloSensitivity × (E − 0.5))
//     λ_away = awayGoalRate × (1 + EloSensitivity × (0.5 − E))
//
// where E = 1 / (10^(−dr/400) + 1) and dr = homeElo − awayElo + homeAdvantage.
//
// Venue adjustment: for a neutral venue homeAdvantage is set to 0; when the
// labelled "home" team is actually playing away the advantage is negated.
//
// Random goal totals are sampled using the Knuth algorithm (exact for λ ≤ 30)
// or a normal approximation (λ > 30, which never occurs in practice for football).
//
// After all iterations the engine returns:
//   - Win/draw/loss probabilities
//   - Expected scoreline (mean goals each side)
//   - Upset probability (probability the lower-Elo team wins)
//   - 95% Wilson-score confidence intervals for each outcome probability
//
// Reference: https://www.eloratings.net/about
package simulator

import (
	"math"
	"math/rand"
	"time"
)

const (
	// BaseGoalRate is the fallback expected-goals-per-game when a team has no
	// match history.  1.25 approximates the historical international-football mean.
	BaseGoalRate = 1.25

	// EloSensitivity (k) controls how strongly the Elo differential shifts λ.
	// At k=1 a team whose Elo predicts a 75 % win chance scores ~1.5× the base
	// rate while their opponent scores ~0.5× the base rate.
	EloSensitivity = 1.0

	// minLambda prevents degenerate Poisson distributions (λ must be > 0).
	minLambda = 0.05

	// DefaultSimulations is used when the caller does not specify a count.
	DefaultSimulations = 1000

	// MaxSimulations caps the iteration count to prevent resource exhaustion.
	MaxSimulations = 10_000
)

// Venue describes the physical venue for the simulated match.
type Venue string

const (
	// VenueHome means the first ("home") team plays at their own ground.
	VenueHome Venue = "home"
	// VenueAway means the first team plays at the second team's ground.
	VenueAway Venue = "away"
	// VenueNeutral means neither team has a home-ground advantage.
	VenueNeutral Venue = "neutral"
)

// Input holds all parameters required for one simulation run.
type Input struct {
	// HomeElo is the current Elo rating of the team labelled "home".
	HomeElo float64
	// AwayElo is the current Elo rating of the team labelled "away".
	AwayElo float64
	// HomeGoalRate is the historical average goals scored per game by the home
	// team.  Pass 0 to fall back to BaseGoalRate.
	HomeGoalRate float64
	// AwayGoalRate is the historical average goals scored per game by the away
	// team.  Pass 0 to fall back to BaseGoalRate.
	AwayGoalRate float64
	// Venue describes where the match is played.
	Venue Venue
	// Simulations is the number of Monte Carlo iterations.  Values ≤ 0 fall
	// back to DefaultSimulations; values > MaxSimulations are capped.
	Simulations int
	// HomeAdvantage is the Elo home-ground bonus used in the expected-result
	// formula.  A value of 0 applies no additional Elo bonus beyond any venue.
	HomeAdvantage float64
}

// Result holds the aggregate outcome of a completed simulation run.
type Result struct {
	// HomeWinPct, DrawPct, AwayWinPct are the estimated probabilities (0–1).
	HomeWinPct float64 `json:"homeWinPct"`
	DrawPct    float64 `json:"drawPct"`
	AwayWinPct float64 `json:"awayWinPct"`

	// ExpectedHomeGoals / ExpectedAwayGoals are the mean goals per iteration.
	ExpectedHomeGoals float64 `json:"expectedHomeGoals"`
	ExpectedAwayGoals float64 `json:"expectedAwayGoals"`

	// UpsetProbability is the probability that the lower-Elo team wins.
	UpsetProbability float64 `json:"upsetProbability"`

	// Simulations is the actual number of iterations executed.
	Simulations int `json:"simulations"`

	// *CI are 95 % Wilson-score confidence intervals for the corresponding
	// probability estimate: [lower, upper].
	HomeWinCI [2]float64 `json:"homeWinCI"`
	DrawCI    [2]float64 `json:"drawCI"`
	AwayWinCI [2]float64 `json:"awayWinCI"`
}

// Run executes the Monte Carlo simulation and returns the aggregate Result.
//
// If rng is nil a new source seeded from the current Unix nanosecond time is
// created automatically, which is suitable for production use.  Pass an
// explicit *rand.Rand seeded with a fixed value in tests for reproducibility.
func Run(input Input, rng *rand.Rand) Result {
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	}

	n := input.Simulations
	switch {
	case n <= 0:
		n = DefaultSimulations
	case n > MaxSimulations:
		n = MaxSimulations
	}

	lambdaHome, lambdaAway := lambdas(input)

	var homeWins, draws, awayWins int
	var totalHomeGoals, totalAwayGoals float64

	for i := 0; i < n; i++ {
		hg := poissonRandom(lambdaHome, rng)
		ag := poissonRandom(lambdaAway, rng)
		totalHomeGoals += float64(hg)
		totalAwayGoals += float64(ag)
		switch {
		case hg > ag:
			homeWins++
		case hg == ag:
			draws++
		default:
			awayWins++
		}
	}

	fn := float64(n)
	homeWinPct := float64(homeWins) / fn
	drawPct := float64(draws) / fn
	awayWinPct := float64(awayWins) / fn

	// Upset probability: probability that the lower-Elo team wins outright.
	var upsetPct float64
	if input.HomeElo >= input.AwayElo {
		upsetPct = awayWinPct
	} else {
		upsetPct = homeWinPct
	}

	return Result{
		HomeWinPct:        roundPct(homeWinPct),
		DrawPct:           roundPct(drawPct),
		AwayWinPct:        roundPct(awayWinPct),
		ExpectedHomeGoals: roundGoals(totalHomeGoals / fn),
		ExpectedAwayGoals: roundGoals(totalAwayGoals / fn),
		UpsetProbability:  roundPct(upsetPct),
		Simulations:       n,
		HomeWinCI:         wilsonCI(homeWinPct, n),
		DrawCI:            wilsonCI(drawPct, n),
		AwayWinCI:         wilsonCI(awayWinPct, n),
	}
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// lambdas computes the Poisson rate parameters (λ_home, λ_away) from the Input.
func lambdas(input Input) (float64, float64) {
	homeRate := input.HomeGoalRate
	if homeRate <= 0 {
		homeRate = BaseGoalRate
	}
	awayRate := input.AwayGoalRate
	if awayRate <= 0 {
		awayRate = BaseGoalRate
	}

	// Resolve the effective home-ground advantage given the venue.
	homeAdv := input.HomeAdvantage
	switch input.Venue {
	case VenueNeutral:
		homeAdv = 0
	case VenueAway:
		// The labelled "home" team is actually playing away: negate the bonus.
		homeAdv = -homeAdv
	}

	E := expectedResult(input.HomeElo, input.AwayElo, homeAdv)

	lambdaHome := homeRate * (1 + EloSensitivity*(E-0.5))
	lambdaAway := awayRate * (1 + EloSensitivity*(0.5-E))

	if lambdaHome < minLambda {
		lambdaHome = minLambda
	}
	if lambdaAway < minLambda {
		lambdaAway = minLambda
	}

	return lambdaHome, lambdaAway
}

// expectedResult computes the Elo expected result (home-win probability).
//
//	E = 1 / (10^(−dr/400) + 1),   dr = homeElo − awayElo + homeAdvantage
func expectedResult(homeElo, awayElo, homeAdvantage float64) float64 {
	dr := homeElo - awayElo + homeAdvantage
	return 1.0 / (math.Pow(10, -dr/400) + 1)
}

// poissonRandom generates a Poisson-distributed random non-negative integer
// with the given mean (lambda) using the Knuth algorithm.  For lambda > 30 a
// normal approximation is substituted to avoid floating-point underflow.
func poissonRandom(lambda float64, rng *rand.Rand) int {
	if lambda <= 0 {
		return 0
	}

	// Knuth algorithm — exact and efficient for small lambda values.
	if lambda <= 30 {
		L := math.Exp(-lambda)
		k := 0
		p := 1.0
		for {
			k++
			p *= rng.Float64()
			if p <= L {
				break
			}
		}
		return k - 1
	}

	// Normal approximation for large lambda (not expected in practice).
	v := rng.NormFloat64()*math.Sqrt(lambda) + lambda
	if v < 0 {
		return 0
	}
	return int(math.Round(v))
}

// wilsonCI returns the 95 % Wilson-score confidence interval [lower, upper]
// for a proportion p estimated from n Bernoulli trials.
func wilsonCI(p float64, n int) [2]float64 {
	if n <= 0 {
		return [2]float64{0, 0}
	}
	const z = 1.96
	fn := float64(n)
	z2 := z * z
	denom := 1 + z2/fn
	center := (p + z2/(2*fn)) / denom
	margin := z * math.Sqrt(p*(1-p)/fn+z2/(4*fn*fn)) / denom
	lo := math.Max(0, center-margin)
	hi := math.Min(1, center+margin)
	return [2]float64{roundPct(lo), roundPct(hi)}
}

// roundPct rounds a probability to four decimal places.
func roundPct(v float64) float64 {
	return math.Round(v*10000) / 10000
}

// roundGoals rounds an expected-goals value to two decimal places.
func roundGoals(v float64) float64 {
	return math.Round(v*100) / 100
}
