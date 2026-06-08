# go-banzhaf

Go implementation of the **Banzhaf power index** for weighted voting games —
built for on-chain token voting, where the input is a token supply and the
distribution of balances across addresses.

[![GoDoc](https://godoc.org/github.com/jbrukh/go-banzhaf?status.svg)](https://godoc.org/github.com/jbrukh/go-banzhaf)

> **v1** is a major rewrite. It adds a family of algorithms (exact and
> approximate), a data-driven dispatcher that picks the best one for your data,
> and a `HighN` routine that handles **hundreds of thousands of holders with raw
> on-chain supplies** in milliseconds. The original `Banzhaf(weights, quota,
> absolute)` function is preserved and unchanged.

## Background

The Banzhaf power index measures voting power in a weighted voting system. A
player is *critical* in a coalition when the coalition wins with them and loses
without them; the index counts these swings. The exact computation is a
subset-sum count and is pseudo-polynomial in the total weight `W` — which is fine
for small governance weights but **impossible for raw on-chain balances**
(an 18-decimal supply is `W ~ 10^27`). For that regime there is a fast,
`W`-independent approximation.

## Quickstart — let the library choose

`Compute` inspects the data and selects the best strategy: the exact method when
the weights are integral and the supply is small enough to afford an array,
otherwise the `W`-independent high-n approximation. Weights are `float64` so raw
on-chain balances (which overflow `uint64`) are accepted directly.

```go
import "github.com/jbrukh/go-banzhaf"

// 100k holders, 18-decimal balances — W ~ 10^25, no problem
res, err := banzhaf.Compute(balances, 0 /* default strict-majority quota */, banzhaf.Options{})
// res.Index    -> normalized power index (sums to 1)
// res.Absolute -> absolute index eta_i / 2^(n-1)
// res.Strategy -> "exact-grouped" or "highn"
// res.Exact    -> true if the result is exact
```

## Algorithms

| Function | Exact? | Complexity | Use |
|----------|--------|-----------|-----|
| `Banzhaf` (legacy) | yes | `O(n·W)` | original per-player exact API |
| `Exact` | yes | `O((n+d)·W)` | exact, grouped by distinct weight |
| `Incremental` | yes | `O(W)` / edit | live dashboards, single-balance edits |
| `NormalApprox` | ~1% | `O(n)` | quick CLT estimate |
| `HighN` | ~ppm–0.1% | `O(n + k·2^k)` | **high n + large raw supply (the target)** |
| `MonteCarlo` | ±ε | `O(s·n)` | absolute indices when some players are pivotal |

`n` = holders, `W` = total supply, `d` = distinct balances, `k` = head size
(`MaxHead`, default 14), `s` = samples.

### The high-n method (`HighN`)

For tens to hundreds of thousands of holders with raw supplies, `HighN` never
touches the weight axis. It splits holders into an exact "head" (the top `k` by
weight, enumerated over all `2^k` subset sums) and a smooth "tail" (handled by the
central limit theorem). A small holder's index is `w_i × density-at-quota`; each
head holder is evaluated exactly by leave-one-out. Cost is `O(n + k·2^k)` and
**completely independent of the supply `W`**.

On scale-free (power-law) holder distributions the head fills to the `MaxHead`
cap and `2^k` dominates the runtime, so `MaxHead` is the speed/accuracy knob.
`k = 14` (the default) is the measured sweet spot. Measured on an Apple M4 Max,
Zipf snapshot, balances × 10^18:

```
  n         HighN time
  100,000    ~18 ms
  250,000    ~33 ms
```

Accuracy vs. the exact method on power-law snapshots: top holders within ~0.02%
for typical concentration (validated in the test suite).

## Cross-language validation

Correctness is pinned to the Python reference implementation that this port was
derived from. `testdata/vectors.json` holds inputs and Python-computed indices;
the Go tests reproduce them:

- `Exact` matches Python's exact normalized **and** absolute index to `< 1e-12`.
- `HighN` matches Python's approximation to `< 1e-9`.
- `Exact` is also cross-checked against an independent `O(2^n)` brute force.

```
go test ./...        # unit, dispatcher, incremental, Monte-Carlo, cross-language
go test -bench .     # HighN throughput
```

## Legacy API

The original function is unchanged and still the simplest exact entry point for
small games:

```go
weights := []uint64{2, 2, 2, 1}
quota := uint64(4)
index, err := banzhaf.Banzhaf(weights, quota, true /* absolute */)
```

Set `banzhaf.ProgressBar = true` for a progress bar on stderr during long exact
runs.

## References

* [Are blockchain voters 'dummies'?](https://blog.coinfund.io/are-blockchain-voters-dummies-4a89a376de69) by @jbrukh
* [Using generator functions to compute power indices](http://www.siue.edu/~aweyhau/teaching/seniorprojects/heger_final.pdf) by Brian Hegers
* https://gist.github.com/HeinrichHartmann/8ec2e2245f2a70441257 by Heinrich Hartmann
