// Package banzhaf computes the Banzhaf power index for weighted voting games,
// with a focus on on-chain token voting (a token supply distributed across many
// addresses).
//
// A player is critical (a swing voter) in a coalition when the coalition wins
// with them and loses without them. The raw Banzhaf score eta_i counts the
// coalitions in which player i is critical; the normalized index is
// eta_i / sum_j eta_j and the absolute index is eta_i / 2^(n-1).
//
// The package offers a family of algorithms and a dispatcher:
//
//   - Compute      data-driven entry point; picks exact or approximate.
//   - Exact        grouped generating-function DP, O((n+d)*W), exact.
//   - Incremental  maintains the generating polynomial for O(W) single edits.
//   - NormalApprox O(n) central-limit estimate of the absolute index.
//   - HighN        head/tail density approximation, O(n + k*2^k), independent of
//     the total weight W — the routine for hundreds of thousands of
//     holders with raw on-chain supplies.
//   - MonteCarlo   sampling estimator with a Hoeffding sample-size helper.
//   - Banzhaf      the original per-player exact API (unchanged).
//
// Exact methods take []uint64 weights; approximate methods and Compute take
// []float64 so that raw on-chain balances (which overflow uint64 and any
// practical array) are accepted directly.
package banzhaf
