//go:build !alt

package pkg

import "math"

func muchBuy(req Request, max, sum float64) float64 {
	// if invested + reserved more than max sum
	// or already collected
	if max-sum <= 0 || req.CollectedPercentage >= 100 {
		return 0
	}

	buy := 0.
	// if more than 30%
	if req.InterestRate >= 0.30 {
		// if sum < 1/6 of max (1000)
		m := max / 6
		if sum < m {
			buy = math.Min(m-sum, m)
		}
		return buy
	}

	// if more than a year
	if req.Term >= 390 {
		// if sum < 1/4 of max (1500)
		m := max / 4
		if sum < m {
			buy = math.Min(m-sum, m)
		}
		return buy
	}

	// if sum < 1/2 of max (3000)
	if sum < max/2 {
		buy = max/2 - sum
	}
	return buy
}
