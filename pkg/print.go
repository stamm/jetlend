package pkg

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
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
	sumTarg := rep.Sum * target
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
			target*100))
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

		if proc >= target+0.0001 {
			if terminal {
				sb.WriteString("\033[31m")
			}
			sb.WriteString(fmt.Sprintf(" > %0.1f%%", target*100))
		}
		if terminal {
			sb.WriteString("\033[0m")
		}
		sb.WriteString("\n")
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
		tm := time.Unix(d.Date/1000, 0)
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
