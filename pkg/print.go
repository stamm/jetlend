package pkg

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	minPercent = 0.20
	// maxInvest  = 6_000
	maxInvest = 0
)

func pr(rep *Report, terminal bool) string {
	var sb strings.Builder
	if !terminal {
		// sb.WriteString("<pre>\n")
		sb.WriteString("```\n")
	}
	p := message.NewPrinter(language.Russian)
	l := float64(len(rep.Values))
	values := make([]float64, 0, len(rep.Values))
	for _, v := range rep.Values {
		values = append(values, v)
	}
	sort.Float64s(values)
	sum := rep.Sum + rep.Reserved + rep.Free

	sb.WriteString(fmt.Sprintf("total    = %s\ninvested = %s (%0.1f%%), len = %0.0f, %0.3f%%\ndelayed  = %s (%0.1f%%)\nreserved = %s (%0.1f%%)\nfree     = %s (%0.1f%%)\nindex    = %0.3f%%, %0.2f, %0.3f\n",
		p.Sprintf("%0.f", sum),
		p.Sprintf("%0.f", rep.Sum), rep.Sum/sum*percent, l, 100./float64(l),
		p.Sprintf("%0.f", rep.Delayed), rep.Delayed/sum*percent,
		p.Sprintf("%0.f", rep.Reserved), rep.Reserved/sum*percent,
		p.Sprintf("%0.f", rep.Free), rep.Free/sum*percent,
		index(rep.Values)*100, math.Sqrt(l*0.02*(1-0.02)), math.Sqrt(0.02*(1-0.02)/l)))

	// sumTarg := sum*target - 500
	sumTarg := rep.Sum * maxTarget
	count := 0
	keys := make([]string, 0, len(rep.Values))
	for key, v := range rep.Values {
		if v >= sum*minTarget && v <= sumTarg {
			count++
		}
		if v <= sumTarg {
			// if v <= 8_500 {
			continue
		}
		keys = append(keys, key)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return rep.Values[keys[i]] < rep.Values[keys[j]]
	})

	sb.WriteString(
		fmt.Sprintf("target: %s (%.2f%%)\ncount: %d\nmax target: %s (%.2f%%)\n",
			p.Sprintf("%05.f", rep.Sum*minTarget),
			minTarget*100,
			count,
			p.Sprintf("%05.f", sumTarg),
			maxTarget*100))
	for _, v := range keys {
		sb.WriteString(p.Sprintf("%s %s \n", p.Sprintf("%05.f", rep.Values[v]), v))
	}
	// for _, q := range []int{50, 75, 90, 95, 96, 97, 98, 99, 100} {
	for _, q := range []float64{50, 75, 90, 95, 99, 100} {
		c := int(math.Round(l*q/percent)) - 1
		if c >= int(l) {
			c = int(l) - 1
		}
		proc := values[c] / rep.Sum

		sb.WriteString(fmt.Sprintf("%3.0fq c=%d(%3d)  %s  %0.2f%%",
			q, c+1, int(l)-c-1, p.Sprintf("%6.f", values[c]), proc*100.0))

		if proc >= maxTarget+0.0001 {
			if terminal {
				sb.WriteString("\033[31m")
			}
			sb.WriteString(fmt.Sprintf(" > %0.1f%%", maxTarget*100))
			// } else if proc >= minTarget+0.0001 {
			// 	if terminal {
			// 		sb.WriteString("\033[31m")
			// 	}
			// 	sb.WriteString(fmt.Sprintf(" > %0.1f%%", minTarget*100))
		}
		if terminal {
			sb.WriteString("\033[0m")
		}
		sb.WriteString("\n")
	}
	sort.Slice(rep.Requests, func(i, j int) bool {
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
		max = int(rep.Sum * minTarget)
	}
	sb.WriteString(fmt.Sprintf("max invest sum: %s\n",
		p.Sprintf("%d", max)))

	t := table.NewWriter()
	t.SetOutputMirror(&sb)
	t.AppendHeader(table.Row{"!", "Buy", "Reserved", "Company", "Percent", "Action", "Sum", "Collected", "Total", "Days", "Rating"})
	for _, req := range rep.Requests {
		if req.InterestRate < minPercent {
			continue
		}
		if req.CollectedPercentage >= 100 {
			continue
		}
		comp := companyNorm(req.LoanName)

		sum := 0.
		if v, ok := rep.Values[comp]; ok {
			sum += v
		}
		sum += req.InvestingAmount.Float64
		// if comp == "СИМБИОЗ" {
		// 	log.Printf("comp %s, req: %+v", comp, req)
		// }
		s := ""

		if math.Abs(float64(max)-sum) < 100 {
			continue
		}

		if int(sum) < max {
			// if terminal {
			// 	sb.WriteString("\033[32m")
			// }
			s = "buy"
		} else if int(sum) > max {
			// if terminal {
			// 	sb.WriteString("\033[31m")
			// }
			s = "sell"
		}
		buy := round(muchBuy(req, float64(max), sum))
		expl := ""
		if max-int(sum) > int(0.75*float64(max)) &&
			req.Term < 390 &&
			req.CollectedPercentage < 100 &&
			req.Rating != "CCC" &&
			req.Amount.Float64 <= 2_500_000 {
			expl = "!"
		}
		if max-int(sum) < int(0.5*float64(max)) {
			expl = "-"
		}
		if buy > 0 && req.CollectedPercentage > 60 && req.CollectedPercentage < 100 {
			expl += "|"
		}
		if s != "" {
			t.AppendRows([]table.Row{
				{expl, buy, req.InvestingAmount.Float64, comp, fmt.Sprintf("%2.1f%%", req.InterestRate*100), s, int(sum),
					int(req.CollectedPercentage),
					fmt.Sprintf("%1.1f", req.Amount.Float64/1_000_000),
					fmt.Sprintf("%d", req.Term),
					req.Rating,
				},
			})
		}
	}
	t.Render()
	if terminal {
		sb.WriteString("\033[0m")
	}

	if !terminal {
		sb.WriteString("```\n")
		// sb.WriteString("</pre>\n")
	}
	return sb.String()
}
func prExpect(exp Expect, terminal bool) string {
	var sb strings.Builder
	p := message.NewPrinter(language.Russian)

	if !terminal {
		sb.WriteString("```\n")
	}

	sum := 0.
	for _, d := range exp.Data {
		sum += d.Amount
		tm := time.Unix(d.Date/1_000, 0)
		sb.WriteString(fmt.Sprintf("%02d: %s\n", tm.Day(),
			p.Sprintf("%7.f", d.Amount)))
	}
	sb.WriteString(fmt.Sprintf("su: %s\n", p.Sprintf("%7.f", sum)))
	if !terminal {
		sb.WriteString("```\n")
		// sb.WriteString("</pre>\n")
	}
	return sb.String()
}

// Herfindahl–Hirschman index
func index(values map[string]float64) float64 {
	sum := 0.
	for _, v := range values {
		sum += v
	}

	s := 0.
	for _, v := range values {
		s += v / sum * v / sum
	}
	return s
}

func muchBuy(req Request, max, sum float64) float64 {
	can := max - sum
	buy := 0.
	// if invested + reserved more than max sum
	// or already collected
	if can <= 0 || req.CollectedPercentage >= 100 {
		return .0
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

func round(buy float64) int {
	if int(buy)%100 > 0 {
		return int(buy) - int(buy)%100
	}
	return int(buy)
}
