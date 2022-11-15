package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	target  = 0.005
	errCode = 401
	percent = 100
)

type Distr struct {
	Data []Data `json:"data"`
}
type Data struct {
	Debt float64 `json:"principal_debt"`
}

type Analytics struct {
	Data DataAn `json:"data"`
}
type DataAn struct {
	Status Status `json:"status"`
}
type Status struct {
	Delayed float64 `json:"delayed"`
}

func main() {
	log.Println("Start")
	var (
		sum    float64
		values []float64
		status Status
		g      errgroup.Group
	)

	g.Go(func() error {
		body, err := getJson("distribution")
		if err != nil {
			return err
		}

		sum, values, err = extractLoans(body)
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		body, err := getJson("analytics")
		if err != nil {
			return err
		}
		status, err = extractAnalitics(body)
		if err != nil {
			return err
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		panic(err)
	}

	pr(sum, values, status)
}

func getJson(uri string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "https://jetlend.ru/invest/api/portfolio/"+uri, nil)
	if err != nil {
		return []byte{}, fmt.Errorf("couldnt create request for uri %s: %w", uri, err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11.15; rv:104.0) Gecko/20100101 Firefox/104.0")
	req.Header.Set("Cookie", os.Getenv("JETLEND_COOKIE"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []byte{}, fmt.Errorf("couldnt do request for uri %s: %w", uri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == errCode {
		return []byte{}, fmt.Errorf("response code %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func extractLoans(body []byte) (float64, []float64, error) {
	var distr Distr
	err := json.Unmarshal(body, &distr)
	if err != nil {
		return 0., []float64{}, fmt.Errorf("cant unmarshal loans: %w", err)
	}

	sum := float64(0)
	values := make([]float64, 0, len(distr.Data))
	for _, d := range distr.Data {
		values = append(values, d.Debt)
		sum += d.Debt
	}
	sort.Float64s(values)
	return sum, values, nil
}

func extractAnalitics(body []byte) (Status, error) {
	var analyt Analytics
	err := json.Unmarshal(body, &analyt)
	if err != nil {
		return Status{}, fmt.Errorf("cant unmarshal analitics: %w", err)
	}

	return analyt.Data.Status, nil
}

func pr(sum float64, values []float64, status Status) {
	l := float64(len(values))
	fmt.Printf("sum = %0.f, len = %0.0f, %0.2f%%\ndelayed = %0.f (%0.1f%%)\n", sum, l, 100./float64(l), status.Delayed, status.Delayed/sum*percent)

	// for _, q := range []int{50, 75, 90, 95, 96, 97, 98, 99, 100} {
	for _, q := range []float64{50, 75, 90, 95, 99, 100} {
		c := int(math.Round(l * q / percent))
		if c >= int(l) {
			c = int(l) - 1
		}
		p := values[c] / sum

		fmt.Printf("%3.0fq c=%d(%3d) %5.f %0.2f%%", q, c+1, int(l)-c, values[c], p*100.0)

		if p > target {
			fmt.Print("\033[31m")
			fmt.Printf(" > %0.1f%%", target*100)
		}
		fmt.Print("\033[0m\n")
	}
}
