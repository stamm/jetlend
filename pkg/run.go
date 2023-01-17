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
	target  = 0.003
	okCode  = 200
	percent = 100
	timeout = 10 * time.Second
)

var ErrGetJson = errors.New("can't get json")

func Run(ctx context.Context, sid string, terminal bool) (string, error) {
	log.Println("Start")
	var (
		rep Report
		g   errgroup.Group
	)

	g.Go(func() error {
		body, err := getJson(ctx, http.DefaultClient, JetUrl("portfolio/distribution"), sid)
		if err != nil {
			return fmt.Errorf("%w for portfolio/distribution (1): %w", ErrGetJson, err)
		}

		rep.Sum, rep.Values, err = extractLoans(body)
		if err != nil {
			return fmt.Errorf("couldn't extract loans for portfolio/distribution (1): %w", err)
		}

		return nil
	})

	g.Go(func() error {
		body, err := getJson(ctx, http.DefaultClient, JetUrl("portfolio/analytics"), sid)
		if err != nil {
			return fmt.Errorf("%w for portfolio/analytics (2): %w", ErrGetJson, err)
		}
		rep.Delayed, err = extractDelayed(body)
		if err != nil {
			return fmt.Errorf("couldn't extract delayed for portfolio/analytics (2): %w", err)
		}

		return nil
	})

	g.Go(func() error {
		body, err := getJson(ctx, http.DefaultClient, JetUrl("account/details"), sid)
		if err != nil {
			return fmt.Errorf("%w for account/details (3): %w", ErrGetJson, err)
		}
		rep.Reserved, rep.Free, err = extractBalance(body)
		if err != nil {
			return fmt.Errorf("couldn't extract balance for account/details (3): %w", err)
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return "", err
	}

	return pr(rep, terminal), nil
}

func getJson(ctx context.Context, client *http.Client, url, sid string) ([]byte, error) {
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

func JetUrl(uri string) string {
	return "https://jetlend.ru/invest/api/" + uri
}
