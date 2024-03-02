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

func needSecondary(rep *Report, terminal, cli bool) (string, bool) {
	var sb strings.Builder
	if !terminal {
		sb.WriteString("```\n")
	}
	// p := message.NewPrinter(language.Russian)

	sort.Slice(rep.Secondary, func(i, j int) bool {
		if rep.Secondary[i].YTM == rep.Secondary[j].YTM {
			return rep.Secondary[i].Company < rep.Secondary[j].Company
		}
		return rep.Secondary[i].YTM > rep.Secondary[j].YTM
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
	for _, sec := range rep.Secondary {

		if sec.TermLeft < 60 || float64(sec.TermLeft)/float64(sec.Term) < 0.3 {
			continue
		}
		if sec.FinancialDiscipline <= 0.9 {
			continue
		}

		if sec.PrincipalDebt <= 200 {
			continue
		}

		if sec.YTM < minPercent {
			continue
			if separate {
				sb.WriteString(strings.Repeat("-", 20) + "\n")
				separate = false
			}
		}

		// log.Printf("sec.InterestRate = %f, sec.CollectedPercentage = %f\n", sec.InterestRate, sec.CollectedPercentage)

		comp := companyNorm(sec.LoanName)

		sum := 0.
		if v, ok := rep.Values[comp]; ok {
			sum += v
		}

		if math.Abs(float64(max)-sum) < 100 {
			continue
		}

		buyFloat, expl := amountBuySecondary(sec, float64(max), sum)
		buy := round(buyFloat)
		if buy <= 0 && expl != "-" {
			continue
		}
		if buy <= 200 {
			continue
		}

		// if expl != "-" {
		// 	continue
		// }
		totalBuy += buy
		count++
		sb.WriteString(fmt.Sprintf("%d, %.0f, %.1f, %s, %s, %d, %d, %d months, %s\n",
			// expl,
			buy, sec.PrincipalDebt,
			sec.MinPrice*100,
			comp,
			fmt.Sprintf("%2.1f%%", sec.YTM*100),
			int(sum),
			sec.TermLeft/30, sec.Term/30,
			sec.Rating,
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
