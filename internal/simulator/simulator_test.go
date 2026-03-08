package simulator_test

import (
	"math"
	"math/rand"
	"testing"

	"github.com/sc23bd/COMP3011_Coursework1/internal/simulator"
)

// deterministicRNG returns a seeded RNG for reproducible tests.
func deterministicRNG() *rand.Rand {
	return rand.New(rand.NewSource(42))
}

// ---------------------------------------------------------------------------
// Run – probability distribution sanity checks
// ---------------------------------------------------------------------------

// TestRun_ProbabilitiesSumToOne verifies that the three outcome probabilities
// always sum to 1.0 (to the precision of four decimal places).
func TestRun_ProbabilitiesSumToOne(t *testing.T) {
	input := simulator.Input{
		HomeElo:     1600,
		AwayElo:     1500,
		Simulations: 5000,
		Venue:       simulator.VenueHome,
	}
	res := simulator.Run(input, deterministicRNG())

	sum := res.HomeWinPct + res.DrawPct + res.AwayWinPct
	if math.Abs(sum-1.0) > 0.0001 {
		t.Errorf("probabilities do not sum to 1.0: got %.4f + %.4f + %.4f = %.4f",
			res.HomeWinPct, res.DrawPct, res.AwayWinPct, sum)
	}
}

// TestRun_FavouriteShouldWinMoreOften checks that the higher-Elo team has a
// higher win probability than the lower-Elo team.
func TestRun_FavouriteShouldWinMoreOften(t *testing.T) {
	input := simulator.Input{
		HomeElo:     1800,
		AwayElo:     1200,
		Simulations: 5000,
		Venue:       simulator.VenueNeutral,
	}
	res := simulator.Run(input, deterministicRNG())

	if res.HomeWinPct <= res.AwayWinPct {
		t.Errorf("expected home (higher Elo) to win more often: home=%.4f, away=%.4f",
			res.HomeWinPct, res.AwayWinPct)
	}
}

// TestRun_EqualTeams_ApproximatelySymmetric verifies that two equal teams on a
// neutral ground produce roughly equal win probabilities.
func TestRun_EqualTeams_ApproximatelySymmetric(t *testing.T) {
	input := simulator.Input{
		HomeElo:     1500,
		AwayElo:     1500,
		Simulations: 10000,
		Venue:       simulator.VenueNeutral,
	}
	res := simulator.Run(input, deterministicRNG())

	diff := math.Abs(res.HomeWinPct - res.AwayWinPct)
	if diff > 0.04 {
		t.Errorf("equal teams should have near-equal win probabilities; diff=%.4f", diff)
	}
}

// TestRun_HomeAdvantageIncreasesHomeWinPct confirms that playing at home
// increases the home team's win probability vs. playing at a neutral venue.
func TestRun_HomeAdvantageIncreasesHomeWinPct(t *testing.T) {
	base := simulator.Input{
		HomeElo:       1500,
		AwayElo:       1500,
		Simulations:   10000,
		HomeAdvantage: 100,
	}

	neutral := base
	neutral.Venue = simulator.VenueNeutral
	resNeutral := simulator.Run(neutral, deterministicRNG())

	home := base
	home.Venue = simulator.VenueHome
	resHome := simulator.Run(home, deterministicRNG())

	if resHome.HomeWinPct <= resNeutral.HomeWinPct {
		t.Errorf("home venue should improve home win pct: home=%.4f neutral=%.4f",
			resHome.HomeWinPct, resNeutral.HomeWinPct)
	}
}

// TestRun_UpsetProbability_Underdog verifies that UpsetProbability equals
// AwayWinPct when the home team has higher Elo.
func TestRun_UpsetProbability_Underdog(t *testing.T) {
	input := simulator.Input{
		HomeElo:     1700,
		AwayElo:     1300,
		Simulations: 2000,
		Venue:       simulator.VenueNeutral,
	}
	res := simulator.Run(input, deterministicRNG())

	if res.UpsetProbability != res.AwayWinPct {
		t.Errorf("upset probability should equal away win pct when home has higher Elo: upset=%.4f away=%.4f",
			res.UpsetProbability, res.AwayWinPct)
	}
}

// TestRun_UpsetProbability_UnderdogIsHome verifies that UpsetProbability equals
// HomeWinPct when the away team has higher Elo.
func TestRun_UpsetProbability_UnderdogIsHome(t *testing.T) {
	input := simulator.Input{
		HomeElo:     1300,
		AwayElo:     1700,
		Simulations: 2000,
		Venue:       simulator.VenueNeutral,
	}
	res := simulator.Run(input, deterministicRNG())

	if res.UpsetProbability != res.HomeWinPct {
		t.Errorf("upset probability should equal home win pct when away has higher Elo: upset=%.4f home=%.4f",
			res.UpsetProbability, res.HomeWinPct)
	}
}

// TestRun_SimulationsCount verifies the Simulations field in the result equals
// the number requested (within the cap).
func TestRun_SimulationsCount(t *testing.T) {
	input := simulator.Input{HomeElo: 1500, AwayElo: 1500, Simulations: 500}
	res := simulator.Run(input, deterministicRNG())
	if res.Simulations != 500 {
		t.Errorf("expected 500 simulations, got %d", res.Simulations)
	}
}

// TestRun_SimulationsCapped verifies that the MaxSimulations cap is enforced.
func TestRun_SimulationsCapped(t *testing.T) {
	input := simulator.Input{HomeElo: 1500, AwayElo: 1500, Simulations: 999999}
	res := simulator.Run(input, deterministicRNG())
	if res.Simulations != simulator.MaxSimulations {
		t.Errorf("expected simulations capped to %d, got %d", simulator.MaxSimulations, res.Simulations)
	}
}

// TestRun_DefaultSimulations verifies that a zero simulation count falls back
// to DefaultSimulations.
func TestRun_DefaultSimulations(t *testing.T) {
	input := simulator.Input{HomeElo: 1500, AwayElo: 1500, Simulations: 0}
	res := simulator.Run(input, deterministicRNG())
	if res.Simulations != simulator.DefaultSimulations {
		t.Errorf("expected default simulations=%d, got %d", simulator.DefaultSimulations, res.Simulations)
	}
}

// TestRun_ExpectedGoalsPositive verifies that expected goals are non-negative.
func TestRun_ExpectedGoalsPositive(t *testing.T) {
	input := simulator.Input{HomeElo: 1500, AwayElo: 1500, Simulations: 1000}
	res := simulator.Run(input, deterministicRNG())
	if res.ExpectedHomeGoals < 0 || res.ExpectedAwayGoals < 0 {
		t.Errorf("expected goals must be non-negative: home=%.2f away=%.2f",
			res.ExpectedHomeGoals, res.ExpectedAwayGoals)
	}
}

// ---------------------------------------------------------------------------
// Confidence interval sanity checks
// ---------------------------------------------------------------------------

// TestWilsonCI_BoundsValid verifies that each CI lower ≤ proportion ≤ upper.
func TestWilsonCI_BoundsValid(t *testing.T) {
	input := simulator.Input{
		HomeElo:     1600,
		AwayElo:     1400,
		Simulations: 5000,
		Venue:       simulator.VenueNeutral,
	}
	res := simulator.Run(input, deterministicRNG())

	check := func(name string, pct float64, ci [2]float64) {
		t.Helper()
		if ci[0] > pct || ci[1] < pct {
			t.Errorf("%s CI [%.4f,%.4f] does not contain proportion %.4f",
				name, ci[0], ci[1], pct)
		}
		if ci[0] > ci[1] {
			t.Errorf("%s CI lower > upper: [%.4f,%.4f]", name, ci[0], ci[1])
		}
	}

	check("homeWin", res.HomeWinPct, res.HomeWinCI)
	check("draw", res.DrawPct, res.DrawCI)
	check("awayWin", res.AwayWinPct, res.AwayWinCI)
}

// TestWilsonCI_InUnitInterval verifies that all CI bounds are in [0, 1].
func TestWilsonCI_InUnitInterval(t *testing.T) {
	input := simulator.Input{HomeElo: 1500, AwayElo: 1500, Simulations: 2000}
	res := simulator.Run(input, deterministicRNG())

	for _, ci := range [][2]float64{res.HomeWinCI, res.DrawCI, res.AwayWinCI} {
		if ci[0] < 0 || ci[1] > 1 {
			t.Errorf("CI out of [0,1]: [%.4f,%.4f]", ci[0], ci[1])
		}
	}
}

// ---------------------------------------------------------------------------
// nil RNG – should not panic
// ---------------------------------------------------------------------------

func TestRun_NilRNG_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Run panicked with nil RNG: %v", r)
		}
	}()
	input := simulator.Input{HomeElo: 1500, AwayElo: 1500, Simulations: 100}
	simulator.Run(input, nil)
}

// ---------------------------------------------------------------------------
// Historical goal rate integration
// ---------------------------------------------------------------------------

// TestRun_HighGoalRate_IncreasesExpectedGoals verifies that a team with a high
// historical scoring rate produces more expected goals than the baseline.
func TestRun_HighGoalRate_IncreasesExpectedGoals(t *testing.T) {
	base := simulator.Input{
		HomeElo: 1500, AwayElo: 1500,
		Simulations: 5000,
		Venue:       simulator.VenueNeutral,
	}
	high := base
	high.HomeGoalRate = 3.0

	resBase := simulator.Run(base, deterministicRNG())
	resHigh := simulator.Run(high, deterministicRNG())

	if resHigh.ExpectedHomeGoals <= resBase.ExpectedHomeGoals {
		t.Errorf("high goal rate should produce more expected goals: high=%.2f base=%.2f",
			resHigh.ExpectedHomeGoals, resBase.ExpectedHomeGoals)
	}
}
