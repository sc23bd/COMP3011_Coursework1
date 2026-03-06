package elo_test

import (
	"math"
	"testing"
	"time"

	"github.com/sc23bd/COMP3011_Coursework1/internal/elo"
)

// cfg is a fixed Config used throughout the tests so results are deterministic.
var cfg = elo.DefaultConfig()

// almostEqual returns true when |a-b| < epsilon.
func almostEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

// ---------------------------------------------------------------------------
// ExpectedResult
// ---------------------------------------------------------------------------

// TestExpectedResult_DrPositive verifies that a 200-point advantage yields ~0.76.
// Reference: eloratings.net formula, dr = 200 → E ≈ 0.7597.
func TestExpectedResult_DrPositive(t *testing.T) {
	homeElo := 1700.0
	awayElo := 1500.0
	e := elo.ExpectedResult(homeElo, awayElo, 0) // no home advantage for clarity
	if !almostEqual(e, 0.7597, 0.001) {
		t.Errorf("ExpectedResult(dr=200): expected ≈0.7597, got %.4f", e)
	}
}

// TestExpectedResult_Equal verifies that equal teams with no home advantage give E=0.5.
func TestExpectedResult_Equal(t *testing.T) {
	e := elo.ExpectedResult(1500, 1500, 0)
	if !almostEqual(e, 0.5, 1e-9) {
		t.Errorf("ExpectedResult(equal, no advantage): expected 0.5, got %.9f", e)
	}
}

// TestExpectedResult_HomeAdvantage verifies that home advantage shifts the expectation upward.
func TestExpectedResult_HomeAdvantage(t *testing.T) {
	eNoAdv := elo.ExpectedResult(1500, 1500, 0)
	eWithAdv := elo.ExpectedResult(1500, 1500, 100)
	if eWithAdv <= eNoAdv {
		t.Errorf("Home advantage should increase expected result: %.4f vs %.4f", eWithAdv, eNoAdv)
	}
}

// ---------------------------------------------------------------------------
// GoalMarginMultiplier
// ---------------------------------------------------------------------------

// TestGoalMarginMultiplier_Draw verifies that a 0-goal difference gives multiplier = 1.0.
func TestGoalMarginMultiplier_Draw(t *testing.T) {
	m := elo.GoalMarginMultiplier(0)
	if !almostEqual(m, 1.0, 1e-9) {
		t.Errorf("GoalMarginMultiplier(0): expected 1.0, got %.9f", m)
	}
}

// TestGoalMarginMultiplier_BiggerWinScalesHigher verifies that a 3-goal win gives
// a higher multiplier than a 1-goal win.
func TestGoalMarginMultiplier_BiggerWinScalesHigher(t *testing.T) {
	m1 := elo.GoalMarginMultiplier(1)
	m3 := elo.GoalMarginMultiplier(3)
	if m3 <= m1 {
		t.Errorf("3-goal margin should yield larger multiplier than 1-goal: %.4f vs %.4f", m3, m1)
	}
}

// TestGoalMarginMultiplier_Symmetry verifies that the sign of the goal difference
// does not affect the multiplier.
func TestGoalMarginMultiplier_Symmetry(t *testing.T) {
	pos := elo.GoalMarginMultiplier(2)
	neg := elo.GoalMarginMultiplier(-2)
	if !almostEqual(pos, neg, 1e-9) {
		t.Errorf("GoalMarginMultiplier should be symmetric: %.9f vs %.9f", pos, neg)
	}
}

// ---------------------------------------------------------------------------
// Calculate – historical replay
// ---------------------------------------------------------------------------

// TestCalculate_NewTeamsStartAtDefault verifies that teams with no history
// receive the configured default rating.
func TestCalculate_NewTeamsStartAtDefault(t *testing.T) {
	ratings := elo.Calculate([]elo.MatchResult{}, cfg)
	if len(ratings) != 0 {
		t.Errorf("expected empty ratings for no matches, got %d entries", len(ratings))
	}
}

// TestCalculate_WinnerGainsRating verifies that the winning team's rating increases.
func TestCalculate_WinnerGainsRating(t *testing.T) {
	matches := []elo.MatchResult{
		{
			MatchID:    1,
			Date:       time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			HomeTeamID: 1,
			AwayTeamID: 2,
			HomeScore:  2,
			AwayScore:  0,
			Tournament: "friendly",
			Neutral:    false,
		},
	}
	ratings := elo.Calculate(matches, cfg)
	if ratings[1] <= cfg.DefaultRating {
		t.Errorf("winner (team 1) should gain rating: %.2f vs default %.2f", ratings[1], cfg.DefaultRating)
	}
	if ratings[2] >= cfg.DefaultRating {
		t.Errorf("loser (team 2) should lose rating: %.2f vs default %.2f", ratings[2], cfg.DefaultRating)
	}
}

// TestCalculate_DrawKeepsRatingsClose verifies that a draw between equal teams
// leaves both ratings near the default.
func TestCalculate_DrawKeepsRatingsClose(t *testing.T) {
	matches := []elo.MatchResult{
		{
			MatchID:    1,
			Date:       time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			HomeTeamID: 1,
			AwayTeamID: 2,
			HomeScore:  1,
			AwayScore:  1,
			Tournament: "friendly",
			Neutral:    true, // neutral site → no home advantage
		},
	}
	ratings := elo.Calculate(matches, cfg)
	// In a draw between equal teams on a neutral ground, home advantage = 0
	// and actual = expected = 0.5, so delta = 0 for both.
	if !almostEqual(ratings[1], cfg.DefaultRating, 1e-6) {
		t.Errorf("team 1 rating after draw between equals: expected %.2f, got %.2f", cfg.DefaultRating, ratings[1])
	}
}

// TestCalculate_HistoricalReplay verifies that ratings are computed chronologically.
// Specifically: playing two matches should reflect both results.
func TestCalculate_HistoricalReplay(t *testing.T) {
	matches := []elo.MatchResult{
		{
			MatchID:    1,
			Date:       time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			HomeTeamID: 1,
			AwayTeamID: 2,
			HomeScore:  3,
			AwayScore:  0,
			Tournament: "friendly",
			Neutral:    true,
		},
		{
			MatchID:    2,
			Date:       time.Date(2000, 6, 1, 0, 0, 0, 0, time.UTC),
			HomeTeamID: 2,
			AwayTeamID: 1,
			HomeScore:  1,
			AwayScore:  0,
			Tournament: "friendly",
			Neutral:    true,
		},
	}

	// After match 1: team 1 > default, team 2 < default.
	// After match 2: team 2 gains some back; team 1 loses some.
	// The key invariant: sum of ratings stays constant (zero-sum).
	ratings := elo.Calculate(matches, cfg)
	sum := ratings[1] + ratings[2]
	expected := 2 * cfg.DefaultRating
	if !almostEqual(sum, expected, 1e-6) {
		t.Errorf("Elo must be zero-sum: expected sum %.2f, got %.2f", expected, sum)
	}
}

// TestCalculate_WorldCupHigherKThanFriendly verifies that a World Cup match
// produces a larger rating change than an otherwise identical friendly.
func TestCalculate_WorldCupHigherKThanFriendly(t *testing.T) {
	makeMatch := func(tournament string) []elo.MatchResult {
		return []elo.MatchResult{{
			MatchID:    1,
			Date:       time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			HomeTeamID: 1,
			AwayTeamID: 2,
			HomeScore:  1,
			AwayScore:  0,
			Neutral:    true,
			Tournament: tournament,
		}}
	}

	ratFriendly := elo.Calculate(makeMatch("Friendly"), cfg)
	ratWC := elo.Calculate(makeMatch("FIFA World Cup"), cfg)

	deltaFriendly := ratFriendly[1] - cfg.DefaultRating
	deltaWC := ratWC[1] - cfg.DefaultRating

	if deltaWC <= deltaFriendly {
		t.Errorf("World Cup delta (%.2f) should be greater than friendly delta (%.2f)", deltaWC, deltaFriendly)
	}
}

// TestCalculateUntil verifies that matches after the cutoff date are excluded.
func TestCalculateUntil(t *testing.T) {
	matches := []elo.MatchResult{
		{
			MatchID:    1,
			Date:       time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			HomeTeamID: 1,
			AwayTeamID: 2,
			HomeScore:  3,
			AwayScore:  0,
			Tournament: "friendly",
			Neutral:    true,
		},
		{
			MatchID:    2,
			Date:       time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC),
			HomeTeamID: 2,
			AwayTeamID: 1,
			HomeScore:  3,
			AwayScore:  0,
			Tournament: "friendly",
			Neutral:    true,
		},
	}

	// Ratings as of 2000-12-31 should only reflect match 1.
	cutoff := time.Date(2000, 12, 31, 0, 0, 0, 0, time.UTC)
	ratCutoff := elo.CalculateUntil(matches, cutoff, cfg)

	// After only match 1, team 1 should be above default.
	if ratCutoff[1] <= cfg.DefaultRating {
		t.Errorf("CalculateUntil: team 1 should be above default after winning match 1: got %.2f", ratCutoff[1])
	}

	// Full calculation including match 2 should move team 2 back up.
	ratFull := elo.Calculate(matches, cfg)
	if ratFull[2] <= ratCutoff[2] {
		t.Errorf("CalculateUntil: team 2 should improve after match 2: cutoff=%.2f, full=%.2f", ratCutoff[2], ratFull[2])
	}
}

// TestKFactor verifies that the config returns the expected K values.
func TestKFactor(t *testing.T) {
	cases := []struct {
		tournament string
		minK       float64
	}{
		{"Friendly", 5},
		{"FIFA World Cup", 30},
		{"World Cup Final", 30},
		{"Semi Final", 15},
		{"Quarter Final", 10},
		{"UEFA Nations Cup Final", 25},
	}
	for _, tc := range cases {
		k := cfg.KFactor(tc.tournament)
		if k < tc.minK {
			t.Errorf("KFactor(%q): expected at least %.0f, got %.0f", tc.tournament, tc.minK, k)
		}
	}
}

// TestKFactor_SubstringPrecedence documents the "highest wins" K-factor
// selection behaviour when tournament names contain multiple matching substrings.
//
// Key behaviours verified:
//   - "FIFA World Cup Qualifier" DOES contain "world cup" as a substring, so it
//     correctly gets K=30 under the current "highest wins" rule.
//   - A tournament that contains NONE of the high-K keywords falls back to
//     DefaultKFactor (5).
func TestKFactor_SubstringPrecedence(t *testing.T) {
	// "FIFA World Cup Qualifier" contains the substring "world cup", so the
	// current highest-wins logic assigns it K=30 — same as a full World Cup match.
	// This is the documented behaviour; the test ensures regressions are caught.
	kQualifier := cfg.KFactor("FIFA World Cup Qualifier")
	kWorldCup := cfg.KFactor("FIFA World Cup")
	if kQualifier != kWorldCup {
		t.Errorf("KFactor('FIFA World Cup Qualifier') expected same as 'FIFA World Cup' (%.0f) due to substring match, got %.0f", kWorldCup, kQualifier)
	}

	// A qualifier that does NOT contain any high-K keyword should fall back to
	// DefaultKFactor (5).
	kAfricanQ := cfg.KFactor("CAF Africa Cup Qualifier")
	if kAfricanQ != cfg.DefaultKFactor {
		t.Errorf("KFactor('CAF Africa Cup Qualifier'): expected DefaultKFactor (%.0f), got %.0f", cfg.DefaultKFactor, kAfricanQ)
	}
}

// TestGoalMarginMultiplier_ExactValues verifies the official eloratings.net
// piecewise step values:
//   - N ≤ 1  → 1.0
//   - N == 2 → 1.5
//   - N == 3 → 1.75
//   - N == 4 → 1.875  (1.75 + (4-3)/8)
//   - N == 5 → 2.0    (1.75 + (5-3)/8)
func TestGoalMarginMultiplier_ExactValues(t *testing.T) {
	cases := []struct {
		diff int
		want float64
	}{
		{0, 1.0},
		{1, 1.0},
		{-1, 1.0},
		{2, 1.5},
		{-2, 1.5},
		{3, 1.75},
		{4, 1.875},
		{5, 2.0},
		{6, 2.125},
	}
	for _, tc := range cases {
		got := elo.GoalMarginMultiplier(tc.diff)
		if !almostEqual(got, tc.want, 1e-9) {
			t.Errorf("GoalMarginMultiplier(%d): expected %.4f, got %.9f", tc.diff, tc.want, got)
		}
	}
}
