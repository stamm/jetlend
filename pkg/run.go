//go:build go1.20

package pkg

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	// target    = 0.0027
	target    = 0.003
	minTarget = 0.002
	okCode    = 200
	percent   = 100
	timeout   = 10 * time.Second
)

// ErrGetJSON is an error getting json
var ErrGetJSON = errors.New("can't get json")

// ExpectAmount show returns for loans
func ExpectAmount(ctx context.Context, sids []string, terminal bool, days int) (string, error) {
	body, err := getJSON(ctx, http.DefaultClient, jetURL(fmt.Sprintf("portfolio/charts/expected_revenue?size=%d", days)), sids[0])
	if err != nil {
		return "", fmt.Errorf("%w for portfolio/charts/expected_revenue: %w", ErrGetJSON, err)
	}
	exp, err := extractExpect(body)
	if err != nil {
		return "", fmt.Errorf("for extract: %w", err)
	}
	return prExpect(exp, terminal), nil
}

func loans(ctx context.Context, rep *Report, sids []string) error {
	var g errgroup.Group
	for _, sid := range sids {
		sid := sid
		g.Go(func() error {
			body, err := getJSON(ctx, http.DefaultClient, jetURL("portfolio/distribution"), sid)
			if err != nil {
				return fmt.Errorf("%w for portfolio/distribution (1): %w", ErrGetJSON, err)
			}

			sum, values, err2 := extractLoans(body)
			if err2 != nil {
				return fmt.Errorf("couldn't extract loans for portfolio/distribution (1): %w", err)
			}

			rep.Mu.Lock()
			rep.Sum += sum
			for k, v := range values {
				if _, ok := rep.Values[k]; ok {
					rep.Values[k] += v
				} else {
					rep.Values[k] = v
				}
			}
			rep.Mu.Unlock()

			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return fmt.Errorf("get error for loans: %s", err)
	}
	return nil
}

func delayed(ctx context.Context, rep *Report, sids []string) error {
	var g errgroup.Group
	for _, sid := range sids {
		sid := sid
		g.Go(func() error {
			body, err := getJSON(ctx, http.DefaultClient, jetURL("portfolio/analytics"), sid)
			if err != nil {
				return fmt.Errorf("%w for portfolio/analytics (2): %w", ErrGetJSON, err)
			}

			delayed, err2 := extractDelayed(body)
			if err2 != nil {
				return fmt.Errorf("couldn't extract delayed for portfolio/analytics (2): %w", err)
			}

			rep.Mu.Lock()
			// log.Printf("delayed: %.0f\n", delayed)
			rep.Delayed += delayed
			rep.Mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("get error for delayed: %s", err)
	}
	return nil
}

func balance(ctx context.Context, rep *Report, sids []string) error {
	var g errgroup.Group
	for _, sid := range sids {
		sid := sid
		g.Go(func() error {
			body, err := getJSON(ctx, http.DefaultClient, jetURL("account/details"), sid)
			if err != nil {
				return fmt.Errorf("%w for account/details (3): %w", ErrGetJSON, err)
			}

			reserved, free, err2 := extractBalance(body)
			if err2 != nil {
				return fmt.Errorf("couldn't extract balance for account/details (3): %w", err)
			}

			rep.Mu.Lock()
			rep.Reserved += reserved
			rep.Free += free
			rep.Mu.Unlock()

			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return fmt.Errorf("get error for balance: %s", err)
	}

	return nil
}

// Run make stats
func Run(ctx context.Context, sids []string, terminal bool) (string, error) {
	log.Println("Start")
	var (
		rep Report
		g   errgroup.Group
	)
	rep.Values = make(map[string]float64, 0)

	g.Go(func() error {
		return loans(ctx, &rep, sids)
	})

	g.Go(func() error {
		return delayed(ctx, &rep, sids)
	})

	g.Go(func() error {
		return balance(ctx, &rep, sids)
	})

	if err := g.Wait(); err != nil {
		return "", err
	}

	return pr(&rep, terminal), nil
}

func getJSON(ctx context.Context, client *http.Client, url, sid string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return []byte{}, fmt.Errorf("couldn't create request for url %s: %w", url, err)
	}

	req.Header.Set("Cookie", "sessionid="+sid)

	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, fmt.Errorf("couldn't do request for url %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != okCode {
		return []byte{}, fmt.Errorf("response code %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func jetURL(uri string) string {
	return "https://jetlend.ru/invest/api/" + uri
}
