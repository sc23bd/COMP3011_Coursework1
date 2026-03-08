# Match Outcome Simulator — Methodology

This document describes the statistical model used by the
`POST /api/v1/football/matches/simulate` endpoint.

---

## Overview

The simulator runs a configurable number of Monte Carlo iterations (default
1 000, maximum 10 000).  Each iteration independently draws a random scoreline
for both teams by sampling from a **Poisson distribution**, then increments one
of three outcome counters (home win / draw / away win).  The final probabilities
are the fraction of iterations that produced each outcome.

---

## Poisson Goal Model

International football goals per team per game follow an approximately Poisson
distribution (Dixon & Coles 1997; Karlis & Ntzoufras 2003).  For a match
between team H (home) and team A (away) the expected-goals parameters are:

```
λ_H = homeGoalRate × (1 + k × (E − 0.5))
λ_A = awayGoalRate × (1 + k × (0.5 − E))
```

| Symbol | Meaning |
|--------|---------|
| `homeGoalRate` | Historical average goals per game for team H (falls back to 1.25) |
| `awayGoalRate` | Historical average goals per game for team A (falls back to 1.25) |
| `E` | Elo expected result for H (probability H wins) — see below |
| `k` | Elo sensitivity constant (default **1.0**) |

The baseline of **1.25 goals per team per game** approximates the long-run
mean in international football results from 1872 to 2025.

---

## Elo Expected Result

The Elo win probability for the home team follows the standard World Football
Elo formula:

```
dr = homeElo − awayElo + homeAdvantage
E  = 1 / (10^(−dr/400) + 1)
```

`homeAdvantage` is taken from the API's Elo configuration (default **100
points**) and is set to zero for neutral-venue matches, or negated when the
labelled "home" team is actually playing away.

---

## Random Variate Generation

Goal totals are generated using the **Knuth algorithm**:

```
L = exp(−λ)
k = 0, p = 1
repeat: k++, p = p × Uniform(0,1)   until p ≤ L
return k − 1
```

This is exact for any λ ≤ 30 (all realistic football λ values).  For λ > 30
(never reached in practice) a normal approximation is substituted.

---

## Venue Factor

| `venue` value | Effect |
|---------------|--------|
| `"home"` | Full `homeAdvantage` applied in dr calculation |
| `"away"` | `homeAdvantage` negated (away team has the ground advantage) |
| `"neutral"` | `homeAdvantage` = 0 |

---

## Confidence Intervals

95 % **Wilson-score** confidence intervals are reported for each outcome
probability.  Wilson intervals are preferred over normal-approximation (Wald)
intervals because they remain valid even when the proportion is near 0 or 1:

```
n        = number of simulations
z        = 1.96  (95 % quantile of standard normal)
center   = (p + z²/(2n)) / (1 + z²/n)
margin   = z × sqrt(p(1−p)/n + z²/(4n²)) / (1 + z²/n)
CI       = [max(0, center−margin),  min(1, center+margin)]
```

---

## Upset Probability

"Upset probability" is defined as the probability that the **lower-Elo team
wins outright**:

- If `homeElo ≥ awayElo`: upset probability = `awayWinPct`  
- If `homeElo < awayElo`: upset probability = `homeWinPct`  

---

## Limitations

1. The model uses only Elo ratings and historical goal averages; it does not
   account for squad availability, recent form streaks, or weather conditions.
2. The Poisson assumption treats both teams' scores as **independent**.
   Real football goals are mildly negatively correlated (scoreline effects
   on play style), but this simplification is standard in the literature for
   pre-match forecasting.
3. The historical goal rate is computed from **all** recorded matches in the
   database, not only recent ones, so it can lag behind changes in team style.

---

## References

- Elo rating formula: https://www.eloratings.net/about  
- Dixon, M. J. & Coles, S. G. (1997). *Modelling association football scores
  and inefficiencies in the football betting market.*  
  Applied Statistics, 46(2), 265–280.  
- Karlis, D. & Ntzoufras, I. (2003). *Analysis of sports data by using bivariate
  Poisson models.*  Journal of the Royal Statistical Society: Series D, 52, 381–393.
