# Elo Methodology

This document describes the dynamic Elo rating system implemented in the
COMP3011 Football API.  It follows the **World Football Elo** method
published at <https://www.eloratings.net/about>.

---

## Formula

```
Elo_new = Elo_old + K × GoalMarginAdj × (ActualResult − ExpectedResult)
```

### Components

| Symbol | Description |
|--------|-------------|
| `K` | Tournament K-factor (see table below) |
| `GoalMarginAdj` | `1 + GoalMarginFactor × ln(\|goal_diff\| + 1)` |
| `ActualResult` | `1.0` = win, `0.5` = draw, `0.0` = loss (from home-team perspective) |
| `ExpectedResult` | `1 / (10^(−dr/400) + 1)` |
| `dr` | `homeElo − awayElo + HomeAdvantage` (set `HomeAdvantage = 0` for neutral sites) |

### Expected Result

The sigmoid logistic curve maps the rating difference to a probability:

```
ExpectedResult = 1 / (10^(−dr/400) + 1)
```

**Example:** `dr = 200` → `ExpectedResult ≈ 0.76`

This means a team that is 200 points stronger is expected to win about 76 %
of the time.

### Goal Margin Adjustment

A larger winning margin earns a proportionally higher rating change:

```
GoalMarginAdj = 1 + GoalMarginFactor × ln(|homeScore − awayScore| + 1)
```

Default `GoalMarginFactor = 0.1` (configurable via `ELO_GOAL_MARGIN_FACTOR`).

| Goal difference | Multiplier (default factor) |
|----------------|-----------------------------|
| 0 (draw)       | 1.00                        |
| 1              | 1.07                        |
| 2              | 1.11                        |
| 3              | 1.14                        |
| 5              | 1.18                        |

### K-Factor Table

The K-factor determines how much weight each match carries.

| Tournament type | K |
|----------------|---|
| World Cup final / World Cup matches | 30 |
| Continental finals (non-WC) | 25 |
| Semi-finals | 15 |
| Quarter-finals | 10 |
| All other (friendlies, qualifiers, group stage, etc.) | 5 |

Selection logic: the system checks if any of the key phrases (`world cup`,
`world cup final`, `final`, `semi`, `quarter`) appear in the tournament name
(case-insensitive) and picks the highest matching K value.

---

## Zero-Sum Property

Elo is a zero-sum system: every point gained by the home team is lost by the
away team and vice versa.  The sum of all ratings across all teams remains
constant over time (equal to `n × DefaultRating` where `n` = number of teams
ever active).

---

## Historical Rating Computation

Ratings are computed by **replaying matches chronologically** from the earliest
available match up to the requested date:

1. All matches up to `?date=YYYY-MM-DD` are fetched from the database, sorted
   ascending by `match_date` then `match_id` (for deterministic ordering of
   same-day matches).
2. Each match is processed in order; both teams' ratings are updated in place.
3. Teams with no prior history start at `DefaultRating = 1500`
   (configurable via `ELO_DEFAULT_RATING`).

This approach guarantees that ratings at date **X** reflect exactly the same
calculation as if you had run the system from scratch on that date.

---

## Home Advantage

When a match is played at a team's home ground, the home team's expected result
is shifted upward by `HomeAdvantage` Elo points (default: 100, configurable via
`ELO_HOME_ADVANTAGE`).  For neutral-site matches (`neutral = true` in the
database) `HomeAdvantage = 0`.

---

## Configuration

All parameters are configurable via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `ELO_DEFAULT_RATING` | `1500` | Starting Elo for teams with no match history |
| `ELO_HOME_ADVANTAGE` | `100` | Rating points added to home-team expected result |
| `ELO_GOAL_MARGIN_FACTOR` | `0.1` | Coefficient for the goal-margin multiplier |

At runtime the `football_elo_config` database table stores the same defaults
and can be used as a source of truth for future UI-driven configuration.

---

## Caching Strategy

Full historical replay across tens of thousands of matches is CPU-intensive.
The `POST /api/v1/football/rankings/elo/recalculate` endpoint (JWT-protected)
triggers a background goroutine that computes ratings for all teams and writes
snapshots to the `football_elo_cache` table.

### Cache-miss behavior

Cached snapshots are read by `GET /rankings/elo?date=YYYY-MM-DD`.  When **no
snapshot exists** for the requested date the rankings endpoint returns an **empty
`data` array** with `X-Cache-Status: miss` response header.  Pre-warm the cache
by calling `POST /rankings/elo/recalculate` before querying rankings.

When the cache **is** populated, the response carries `X-Cache-Status: hit`.

Individual team Elo queries (`GET /teams/:id/elo`) always compute on-demand
from the live `football_matches` data, so they are always accurate without
requiring a pre-warm step.

### Rate limiting on `/recalculate`

To prevent accidental or malicious DoS via repeated recalculation requests, the
endpoint enforces the following constraints:

| Condition | Response |
|-----------|----------|
| Recalculation is already running | `429 Too Many Requests` |
| Last run completed < 5 minutes ago | `429 Too Many Requests` |
| `?force=true` supplied | Bypasses the 5-minute cooldown |

All recalculation activity (start, finish, errors) is logged at `INFO` level
with duration so administrators can monitor performance.

---

## Limitations

* **Cache staleness** — rankings snapshots (`GET /rankings/elo`) are computed at
  a point in time.  Use `POST /recalculate` to refresh them.  Individual team
  Elo queries are always computed live from the match database.
* **No player-level weighting** — the model treats every outfield player
  identically; substitutions, injuries, and squad quality are not considered.
* **Friendly match weighting** — the low K-factor (5) means friendlies barely
  move ratings, which is intentional but may not reflect true team strength
  during periods when a team plays only friendlies.
* **Name changes** — historical team name changes are tracked in
  `football_former_names` but the Elo engine currently uses the current team ID
  throughout, so pre-rename and post-rename ratings are correctly combined.
* **Tournament classification** — K-factor selection relies on substring
  matching of tournament names.  Edge cases (e.g. "Euro 2020 Final") are
  handled by the `final` key (K=25), but unusual tournament naming may fall
  through to the default K=5.

---

## References

* Arpad Elo (1978). *The Rating of Chessplayers, Past and Present*. Arco Publishing.
* World Football Elo Ratings method: <https://www.eloratings.net/about>
* Kaggle dataset: [International football results 1872–2025](https://www.kaggle.com/datasets/martj42/international-football-results-from-1872-to-2017)
