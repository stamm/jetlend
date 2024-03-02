package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/stamm/jetlend/pkg"
)

const (
	okCode     = http.StatusOK
	defaultURL = "https://jetlend.ru/invest/api"
)

// ErrGetJSON is an error getting json
var ErrGetJSON = errors.New("can't get json")

type Jetlend struct {
	Timeout time.Duration
	URL     string
}

var _ HTTP = (*Jetlend)(nil)

func NewJetlend(url string, timeout time.Duration) *Jetlend {
	if url == "" {
		url = defaultURL
	}
	return &Jetlend{
		URL:     url,
		Timeout: timeout,
	}
}

func (j *Jetlend) getJSON(ctx context.Context, client *http.Client, url, sid string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, j.Timeout)
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

func (j Jetlend) url(uri string) string {
	return j.URL + "/" + uri
}

func (j Jetlend) urlExpect(days int) string {
	return j.url(fmt.Sprintf("portfolio/charts/expected_revenue?size=%d", days))
}

// ExpectAmount show returns for loans
func (j Jetlend) ExpectAmount(ctx context.Context, cfg pkg.Config, days int) ([]byte, error) {
	if len(cfg.Sids) == 0 {
		return []byte{}, errors.New("no config")
	}
	body, err := j.getJSON(ctx,
		http.DefaultClient,
		j.urlExpect(days),
		cfg.Sids[0])
	if err != nil {
		return []byte{}, fmt.Errorf("%w for portfolio/charts/expected_revenue: %w", ErrGetJSON, err)
	}
	return body, nil
}

// func (j Jetlend) LoansPorfolio(ctx context.Context, cfg pkg.Config) error {
// 	var g errgroup.Group
// 	for _, sid := range sids {
// 		sid := sid
// 		g.Go(func() error {
// 			body, err := getJSON(ctx, http.DefaultClient, jetURL("portfolio/distribution"), sid)
// 			if err != nil {
// 				return fmt.Errorf("%w for portfolio/distribution (1): %w", ErrGetJSON, err)
// 			}

// 			sum, values, err2 := extractLoans(body)
// 			if err2 != nil {
// 				return fmt.Errorf("couldn't extract loans for portfolio/distribution (1): %w", err2)
// 			}

// 			rep.Mu.Lock()
// 			rep.Sum += sum
// 			for k, v := range values {
// 				if _, ok := rep.Values[k]; ok {
// 					rep.Values[k] += v
// 				} else {
// 					rep.Values[k] = v
// 				}
// 			}
// 			rep.Mu.Unlock()

// 			return nil
// 		})
// 	}
// 	if err := g.Wait(); err != nil {
// 		return fmt.Errorf("get error for loans: %s", err)
// 	}
// 	return nil
// }

// func (j Jetlend) Delayed(ctx context.Context, cfg pkg.Config) error {
// 	var g errgroup.Group
// 	for _, sid := range sids {
// 		sid := sid
// 		g.Go(func() error {
// 			body, err := getJSON(ctx, http.DefaultClient, jetURL("portfolio/analytics"), sid)
// 			if err != nil {
// 				return fmt.Errorf("%w for portfolio/analytics (2): %w", ErrGetJSON, err)
// 			}

// 			delayed, err2 := extractDelayed(body)
// 			if err2 != nil {
// 				return fmt.Errorf("couldn't extract delayed for portfolio/analytics (2): %w", err2)
// 			}

// 			rep.Mu.Lock()
// 			// log.Printf("delayed: %.0f\n", delayed)
// 			rep.Delayed += delayed
// 			rep.Mu.Unlock()

// 			return nil
// 		})
// 	}

// 	if err := g.Wait(); err != nil {
// 		return fmt.Errorf("get error for delayed: %s", err)
// 	}
// 	return nil
// }

// func (j Jetlend) Balance(ctx context.Context, rep *Report, sids []string) error {
// 	var g errgroup.Group
// 	for _, sid := range sids {
// 		sid := sid
// 		g.Go(func() error {
// 			body, err := getJSON(ctx, http.DefaultClient, jetURL("account/details"), sid)
// 			if err != nil {
// 				return fmt.Errorf("%w for account/details (3): %w", ErrGetJSON, err)
// 			}

// 			reserved, free, err2 := extractBalance(body)
// 			if err2 != nil {
// 				return fmt.Errorf("couldn't extract balance for account/details (3): %w", err2)
// 			}

// 			rep.Mu.Lock()
// 			rep.Reserved += reserved
// 			rep.Free += free
// 			rep.Mu.Unlock()

// 			return nil
// 		})
// 	}
// 	if err := g.Wait(); err != nil {
// 		return fmt.Errorf("get error for balance: %s", err)
// 	}

// 	return nil
// }

// func (j Jetlend) PrimaryMarket(ctx context.Context, rep *Report, sid string) error {
// 	body, err := getJSON(ctx, http.DefaultClient, jetURL("requests/waiting"), sid)
// 	if err != nil {
// 		return fmt.Errorf("%w for requests/waiting: %w", ErrGetJSON, err)
// 	}

// 	requests, err2 := extractRequests(body)
// 	if err2 != nil {
// 		return fmt.Errorf("couldn't extract requests for requests/waiting: %w", err2)
// 	}

// 	rep.Mu.Lock()
// 	rep.Requests = requests
// 	rep.Mu.Unlock()

// 	return nil
// }

// func (j Jetlend) SecondaryMarket(ctx context.Context, rep *Report, sid string) error {
// 	body, err := getJSON(ctx, http.DefaultClient, jetURL("exchange/loans?limit=1000&offset=0&sort_dir=desc&sort_field=ytm"), sid)
// 	if err != nil {
// 		return fmt.Errorf("%w for exchange/loans: %w", ErrGetJSON, err)
// 	}
// 	// fmt.Printf("body: %s", string(body))

// 	secondary, err2 := extractSecondary(body)
// 	if err2 != nil {
// 		return fmt.Errorf("couldn't extract requests for exchange/loans: %w", err2)
// 	}

// 	rep.Mu.Lock()
// 	rep.Secondary = secondary
// 	rep.Mu.Unlock()

// 	return nil
// }
