package pkg

import (
	"fmt"
	"math"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func pr(rep Report, terminal bool) string {
	var sb strings.Builder
	if !terminal {
		// sb.WriteString("<pre>\n")
		sb.WriteString("```\n")
	}
	p := message.NewPrinter(language.Russian)
	l := float64(len(rep.Values))
	sum := rep.Sum + rep.Reserved + rep.Free
	sb.WriteString(fmt.Sprintf("invested = %s, len = %0.0f, %0.3f%%\ndelayed = %s (%0.1f%%)\nreserved = %s (%0.1f%%)\nfree = %s (%0.1f%%)\nindex = %0.2f%%\n",
		p.Sprintf("%0.f", rep.Sum), l, 100./float64(l),
		p.Sprintf("%0.f", rep.Delayed), rep.Delayed/sum*percent,
		p.Sprintf("%0.f", rep.Reserved), rep.Reserved/sum*percent,
		p.Sprintf("%0.f", rep.Free), rep.Free/sum*percent,
		index(rep.Values)*100))

	// for _, q := range []int{50, 75, 90, 95, 96, 97, 98, 99, 100} {
	for _, q := range []float64{50, 75, 90, 95, 99, 100} {
		c := int(math.Round(l*q/percent)) - 1
		if c >= int(l) {
			c = int(l) - 1
		}
		proc := rep.Values[c] / sum

		sb.WriteString(fmt.Sprintf("%3.0fq c=%d(%3d)  %s  %0.2f%%",
			q, c+1, int(l)-c-1, p.Sprintf("%6.f", rep.Values[c]), proc*100.0))

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

// Herfindahl–Hirschman index
func index(values []float64) float64 {
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
