package pkg

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	target  = 0.005
	okCode  = 200
	percent = 100
	timeout = 5 * time.Second
)

func Run() {
	log.Println("Start")
	var (
		rep Report
		g   errgroup.Group
	)

	g.Go(func() error {
		body, err := getJson("portfolio/distribution")
		if err != nil {
			return err
		}

		rep.Sum, rep.Values, err = extractLoans(body)
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		body, err := getJson("portfolio/analytics")
		if err != nil {
			return err
		}
		rep.Delayed, err = extractDelayed(body)
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		body, err := getJson("account/details")
		if err != nil {
			return err
		}
		rep.Reserved, err = extractReserved(body)
		if err != nil {
			return err
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		panic(err)
	}

	pr(rep)
}

func getJson(uri string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://jetlend.ru/invest/api/"+uri, nil)
	if err != nil {
		return []byte{}, fmt.Errorf("couldnt create request for uri %s: %w", uri, err)
	}

	req.Header.Set("Cookie", os.Getenv("JETLEND_COOKIE"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []byte{}, fmt.Errorf("couldnt do request for uri %s: %w", uri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != okCode {
		return []byte{}, fmt.Errorf("response code %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func pr(rep Report) {
	p := message.NewPrinter(language.Russian)
	l := float64(len(rep.Values))
	fmt.Printf("sum = %s, len = %0.0f, %0.2f%%\ndelayed = %s (%0.1f%%)\nreserved = %s (%0.1f%%)\n",
		p.Sprintf("%0.f", rep.Sum), l, 100./float64(l),
		p.Sprintf("%0.f", rep.Delayed), rep.Delayed/rep.Sum*percent,
		p.Sprintf("%0.f", rep.Reserved), rep.Reserved/rep.Sum*percent)

	// for _, q := range []int{50, 75, 90, 95, 96, 97, 98, 99, 100} {
	for _, q := range []float64{50, 75, 90, 95, 99, 100} {
		c := int(math.Round(l * q / percent))
		if c >= int(l) {
			c = int(l) - 1
		}
		proc := rep.Values[c] / rep.Sum

		fmt.Printf("%3.0fq c=%d(%3d)  %s  %0.2f%%",
			q, c+1, int(l)-c, p.Sprintf("%6.f", rep.Values[c]), proc*100.0)

		if proc > target {
			fmt.Print("\033[31m")
			fmt.Printf(" > %0.1f%%", target*100)
		}
		fmt.Print("\033[0m\n")
	}
}
