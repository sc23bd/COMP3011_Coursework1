// Package elo provides Elo rating calculation and configuration for the
// dynamic international football Elo rating system.
package elo

import (
	"os"
	"strconv"
	"strings"
)

const (
	formulaReference = "https://www.eloratings.net/method.html"
)

// Config holds all tunable parameters for the Elo calculation engine.
// Values are loaded from environment variables with sensible defaults.
type Config struct {
	// DefaultRating is the starting Elo for teams with no match history.
	DefaultRating float64
	// HomeAdvantage is the number of Elo points added to the home team's
	// expected-result calculation.  Set to 0 for neutral-site matches.
	HomeAdvantage float64
	// GoalMarginFactor is the coefficient applied to ln(|goal_diff|+1) when
	// scaling the K factor for goal-margin adjustment.
	GoalMarginFactor float64
	// DefaultKFactor is the fallback K used when no tournament-name rule matches.
	DefaultKFactor float64
	// KFactorRules maps lower-cased tournament name substrings to their K value.
	// The calculator tests each key as a substring of the tournament name (case-
	// insensitive); the highest-matching K value wins.
	KFactorRules map[string]float64
}

// DefaultConfig returns a Config pre-loaded from environment variables,
// falling back to the standard World Football Elo parameters.
//
// Environment variables recognised:
//
//	ELO_DEFAULT_RATING    – float, default 1500
//	ELO_HOME_ADVANTAGE    – float, default 100
//	ELO_GOAL_MARGIN_FACTOR – float, default 0.1
func DefaultConfig() Config {
	cfg := Config{
		DefaultRating:    parseEnvFloat("ELO_DEFAULT_RATING", 1500),
		HomeAdvantage:    parseEnvFloat("ELO_HOME_ADVANTAGE", 100),
		GoalMarginFactor: parseEnvFloat("ELO_GOAL_MARGIN_FACTOR", 0.1),
		// DefaultKFactor applies to friendlies, qualifiers, and any tournament
		// that does not match a more specific rule below.
		DefaultKFactor: 5,
		KFactorRules: map[string]float64{
			// World Cup final and knockout stages
			"world cup final": 30,
			"world cup":       30,
			// Continental finals (non-WC)
			"final": 25,
			// Semi-finals
			"semi": 15,
			// Quarter-finals
			"quarter": 10,
		},
	}
	return cfg
}

// KFactor returns the K value for the given tournament name.
// The method selects the entry whose key is a substring of the (lower-cased)
// tournament name and yields the highest K value.
// Falls back to DefaultKFactor when no rule matches.
func (c Config) KFactor(tournament string) float64 {
	lower := strings.ToLower(tournament)
	best := c.DefaultKFactor
	for key, k := range c.KFactorRules {
		if strings.Contains(lower, key) && k > best {
			best = k
		}
	}
	return best
}

// FormulaRef returns the public reference URL for the Elo method.
func (c Config) FormulaRef() string {
	return formulaReference
}

// parseEnvFloat reads an environment variable as a float64.
// Returns def when the variable is unset or cannot be parsed.
func parseEnvFloat(key string, def float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
}
