package pkg

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"log/slog"

	"golang.org/x/sync/errgroup"
)

const (
	// target    = 0.0027
	minTargetPercent = 0.00206
	minTargetSum     = 3_901.
	maxTargetPercent = 0.003
	maxTargetSum     = 4_001.
	okCode           = 200
	percent          = 100
	timeout          = 20 * time.Second
)

type Config struct {
	Sids     []string
	Terminal bool
}
type HTTP interface {
	ExpectAmount(ctx context.Context, cfg Config, days int) (string, error)
	LoansPorfolio(ctx context.Context, cfg Config) (string, error)
	Delayed(ctx context.Context, cfg Config) (int, error)
	Balance(ctx context.Context, cfg Config) (int, int, error)
	PrimaryMarket(ctx context.Context, cfg Config) (string, error)
	SecondaryMarket(ctx context.Context, cfg Config) (string, error)
}

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
				return fmt.Errorf("couldn't extract loans for portfolio/distribution (1): %w", err2)
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
				return fmt.Errorf("couldn't extract delayed for portfolio/analytics (2): %w", err2)
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
				return fmt.Errorf("couldn't extract balance for account/details (3): %w", err2)
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

func requests(ctx context.Context, rep *Report, sid string) error {
	body, err := getJSON(ctx, http.DefaultClient, jetURL("requests/waiting"), sid)
	if err != nil {
		return fmt.Errorf("%w for requests/waiting: %w", ErrGetJSON, err)
	}

	requests, err2 := extractRequests(body)
	if err2 != nil {
		return fmt.Errorf("couldn't extract requests for requests/waiting: %w", err2)
	}

	rep.Mu.Lock()
	rep.Requests = requests
	rep.Mu.Unlock()
	// fmt.Printf("%+v", requests)

	return nil
}

func secondary(ctx context.Context, rep *Report, sid string) error {
	all := make([]Secondary, 0)

	offset := 0
	limit := 100
	for {
		uri := fmt.Sprintf("exchange/loans?limit=%d&offset=%d&sort_dir=desc&sort_field=ytm", limit, offset)
		slog.Info(uri)
		body, err := getJSON(ctx, http.DefaultClient, jetURL(
			uri,
		), sid)
		if err != nil {
			return fmt.Errorf("%w for exchange/loans: %w", ErrGetJSON, err)
		}
		// fmt.Printf("body: %s", string(body))

		secondary, total, err2 := extractSecondary(body)
		if err2 != nil {
			return fmt.Errorf("couldn't extract requests for exchange/loans: %w", err2)
		}
		all = append(all, secondary...)
		if offset+limit > total || secondary[0].YTM < 0.18 {
			break
		}
		offset += limit
	}

	rep.Mu.Lock()
	rep.Secondary = all
	rep.Mu.Unlock()

	return nil
}

// Run make stats
func Run(ctx context.Context, sids []string, terminal, cli bool) (string, error) {
	slog.Info("Start")
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

	g.Go(func() error {
		sid := sids[0]
		if len(sids) > 1 {
			sid = sids[1]
		}
		return requests(ctx, &rep, sid)
	})

	// g.Go(func() error {
	// 	sid := sids[0]
	// 	if len(sids) > 1 {
	// 		sid = sids[1]
	// 	}
	// 	return secondary(ctx, &rep, sid)
	// })

	if err := g.Wait(); err != nil {
		return "", err
	}

	return pr(&rep, terminal, cli), nil
}

// WhatBuy every hour
func WhatBuy(ctx context.Context, sids []string, terminal, cli bool) (string, bool, error) {
	slog.Info("Start what to buy")
	var (
		rep Report
		g   errgroup.Group
	)
	rep.Values = make(map[string]float64, 0)

	g.Go(func() error {
		return loans(ctx, &rep, sids)
	})

	g.Go(func() error {
		sid := sids[0]
		if len(sids) > 1 {
			sid = sids[1]
		}
		return requests(ctx, &rep, sid)
	})

	if err := g.Wait(); err != nil {
		return "", false, err
	}

	s, have := needAction(&rep, terminal, cli)
	return s, have, nil
}

// SecondaryMarket
func SecondaryMarket(ctx context.Context, sids []string, terminal, cli bool) (string, bool, error) {
	slog.Info("Start what to buy")
	var (
		rep Report
		g   errgroup.Group
	)
	rep.Values = make(map[string]float64, 0)

	g.Go(func() error {
		return loans(ctx, &rep, sids)
	})

	// g.Go(func() error {
	// 	sid := sids[0]
	// 	if len(sids) > 1 {
	// 		sid = sids[1]
	// 	}
	// 	return requests(ctx, &rep, sid)
	// })

	g.Go(func() error {
		sid := sids[0]
		if len(sids) > 1 {
			sid = sids[1]
		}
		return secondary(ctx, &rep, sid)
	})

	if err := g.Wait(); err != nil {
		return "", false, err
	}

	s, have := needSecondary(&rep, terminal, cli)
	return s, have, nil
}

func getJSON(ctx context.Context, client *http.Client, url, sid string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return []byte{}, fmt.Errorf("couldn't create request for url %s: %w", url, err)
	}

	req.Header.Set("Cookie", "sessionid="+sid)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:109.0) Gecko/20100101 Firefox/114.0")

	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, fmt.Errorf("couldn't do request for url %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != okCode {
		b, _ := io.ReadAll(resp.Body)
		return []byte{}, fmt.Errorf("response code %d: %s", resp.StatusCode, string(b))
	}

	return io.ReadAll(resp.Body)
}

func jetURL(uri string) string {
	return "https://jetlend.ru/invest/api/" + uri
}
