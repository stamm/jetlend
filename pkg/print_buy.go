package pkg

import (
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

func needAction(rep *Report, terminal, cli bool) (string, bool) {
	var sb strings.Builder
	if !terminal {
		sb.WriteString("```\n")
	}
	// p := message.NewPrinter(language.Russian)

	sort.Slice(rep.Requests, func(i, j int) bool {
		if rep.Requests[i].InterestRate == rep.Requests[j].InterestRate {
			return rep.Requests[i].Company < rep.Requests[j].Company
		}
		return rep.Requests[i].InterestRate > rep.Requests[j].InterestRate
	})

	maxStr, ok := os.LookupEnv("JETLEND_MAX")
	if !ok {
		maxStr = "0"
	}
	max, err := strconv.Atoi(maxStr)
	if err != nil {
		panic(err)
	}
	if max == 0 {
		max = int(minTargetSum)
	}

	totalBuy := 0
	count := 0
	separate := true
	for _, req := range rep.Requests {
		if req.InterestRate < minPercent {
			if separate {
				sb.WriteString(strings.Repeat("-", 20) + "\n")
				separate = false
			}
		}

		// log.Printf("req.InterestRate = %f, req.CollectedPercentage = %f\n", req.InterestRate, req.CollectedPercentage)
		if req.CollectedPercentage >= 100 {
			continue
		}

		comp := companyNorm(req.LoanName)

		sum := 0.
		if v, ok := rep.Values[comp]; ok {
			sum += v
		}
		sum += req.InvestingAmount.Float64

		if math.Abs(float64(max)-sum) < 100 {
			continue
		}

		buyFloat, expl := muchBuy(req, float64(max), sum)
		buy := round(buyFloat)
		if buy <= 0 && expl != "-" {
			continue
		}
		if expl != "-" && req.CollectedPercentage <= 80 {
			continue
		}
		totalBuy += buy
		count++
		sb.WriteString(fmt.Sprintf("%s, %d, %.0f, %s, %s, %d, %d%%, %s mln, %s months, %s\n",
			expl,
			buy,
			req.InvestingAmount.Float64,
			comp, fmt.Sprintf("%2.1f%%", req.InterestRate*100), int(sum),
			int(req.CollectedPercentage),
			fmt.Sprintf("%1.1f", req.Amount.Float64/1_000_000),
			fmt.Sprintf("%d", req.Term/30),
			req.Rating,
		))
	}
	sb.WriteString(fmt.Sprintf("total to buy : %d\n", totalBuy))
	if terminal {
		sb.WriteString("\033[0m")
	}

	if !terminal {
		sb.WriteString("```\n")
	}
	log.Printf("count = %d\n", count)
	return sb.String(), count > 0
}
