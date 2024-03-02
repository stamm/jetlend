//go:build !alt

package pkg

import "math"

func muchBuy(req Request, max, sum float64) (float64, string) {
	expl := ""

	if (max - sum) < max {
		expl = "✓"
		if req.InvestingAmount.Float64 > 0 {
			expl = "-"
		}
	}
	if req.CollectedPercentage < 85 && req.InvestingAmount.Float64 > 0 {
		expl = "-"
	}
	// if invested + reserved more than max sum
	// or already collected
	if max-sum <= 0 || req.CollectedPercentage >= 100 {
		return 0, expl
	}

	buy := 0.
	// if more than 30%
	if req.InterestRate >= 0.30 {
		// if sum < 1/2 of max (1500)
		m := max / 2
		if sum < m {
			buy = math.Min(m-sum, m)
		}
		return buy, mark(req, max, sum, buy, expl)
	}

	// if more than a year
	if req.Term >= 390 {
		// if sum < 1/2 of max (1500)
		if sum < max {
			buy = math.Min(max-sum, max)
		}
		return buy, mark(req, max, sum, buy, expl)
	}

	// if sum < max (3000)
	if sum < max {
		buy = max - sum
	}
	return buy, mark(req, max, sum, buy, expl)
}
func amountBuySecondary(sec Secondary, max, sum float64) (float64, string) {
	buy := 0.
	// if more than 30%
	if sec.YTM >= 0.30 {
		// if sum < 1/6 of max (1500)
		m := max / 2
		if sum < m {
			buy = math.Min(m-sum, m)
		}
		return buy, ""
	}

	// if more than a year
	if sec.Term >= 390 {
		// if sum < 1/4 of max (1500)
		m := max / 2
		if sum < m {
			buy = math.Min(m-sum, m)
		}
		return buy, ""
	}

	// if sum < max (3000)
	if sum < max {
		buy = max - sum
	}
	return buy, ""
}

func mark(req Request, max, sum, buy float64, expl string) string {
	if buy < 100 || req.CollectedPercentage < 60 || req.CollectedPercentage > 100 {
		return expl
	}
	expl += "░"
	if req.CollectedPercentage > 70 {
		expl += "▒"
	}
	if req.CollectedPercentage > 80 {
		expl += "▓"
	}
	if req.CollectedPercentage > 90 {
		expl += "█"
	}
	return expl

}
