//go:build !alt

package pkg

import "math"

func muchBuy(req Request, max, sum float64) float64 {
	can := max - sum
	buy := 0.
	// if invested + reserved more than max sum
	// or already collected
	if can <= 0 || req.CollectedPercentage >= 100 {
		return buy
	}

	// if more than 30%
	if req.InterestRate >= 0.30 {
		// if sum < 1/6 of max (1000)
		m := max / 6
		if sum < float64(m) {
			buy = math.Min(float64(can), float64(m))
		}
		return buy
	}

	// if more than a year
	if req.Term >= 390 {
		// if sum < 1/4 of max (1500)
		if sum < float64(max/4) {
			buy = math.Min(float64(can), float64(max/4))
		}
		return buy
	}

	// if sum < 1/2 of max (3000)
	if sum < float64(max/2) {
		buy = float64(max/2) - sum
	}
	return buy
}
