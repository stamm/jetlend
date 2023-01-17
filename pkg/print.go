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
	sb.WriteString(fmt.Sprintf("invested = %s, len = %0.0f, %0.3f%%\ndelayed = %s (%0.1f%%)\nreserved = %s (%0.1f%%)\nfree = %s (%0.1f%%)\n",
		p.Sprintf("%0.f", rep.Sum), l, 100./float64(l),
		p.Sprintf("%0.f", rep.Delayed), rep.Delayed/sum*percent,
		p.Sprintf("%0.f", rep.Reserved), rep.Reserved/sum*percent,
		p.Sprintf("%0.f", rep.Free), rep.Free/sum*percent))

	// for _, q := range []int{50, 75, 90, 95, 96, 97, 98, 99, 100} {
	for _, q := range []float64{50, 75, 90, 95, 99, 100} {
		c := int(math.Round(l * q / percent))
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
