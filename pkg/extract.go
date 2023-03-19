//go:build go1.20

package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var ErrUnmarshal = errors.New("can't unmarshal")

func extractExpect(body []byte) (Expect, error) {
	var expect Expect
	err := json.Unmarshal(body, &expect)
	if err != nil {
		return Expect{}, fmt.Errorf("%w expect: %w", ErrUnmarshal, err)
	}
	return expect, nil
}

func extractLoans(body []byte) (float64, map[string]float64, error) {
	var distr Distr
	err := json.Unmarshal(body, &distr)
	if err != nil {
		return 0., map[string]float64{}, fmt.Errorf("%w loans: %w", ErrUnmarshal, err)
	}

	sum := float64(0)
	comp := make(map[string]float64, len(distr.Data))
	count := make(map[string]int)
	for _, d := range distr.Data {
		com := companyNorm(d.Company)
		if comp[com] != 0 {
			// 	log.Printf("spl[0] = %+v\n", spl[0])
			if _, ok := count[com]; !ok {
				count[com] = 2
			} else {
				count[com]++
			}
		}
		comp[com] += d.Debt
	}
	// log.Printf("count = %+v\n", count)
	for _, v := range comp {
		sum += v
	}
	return sum, comp, nil
}

func extractDelayed(body []byte) (float64, error) {
	var analyt Analytics
	err := json.Unmarshal(body, &analyt)
	if err != nil {
		return .0, fmt.Errorf("%w analitics: %w", ErrUnmarshal, err)
	}

	return analyt.Data.Status.Delayed, nil
}

func extractBalance(body []byte) (float64, float64, error) {
	var details Details
	err := json.Unmarshal(body, &details)
	if err != nil {
		return .0, .0, fmt.Errorf("%w analitics: %w", ErrUnmarshal, err)
	}

	return details.Data.Balance.Reserved, details.Data.Balance.Free, nil
}

func extractRequests(body []byte) ([]Request, error) {
	var waiting Waiting
	err := json.Unmarshal(body, &waiting)
	if err != nil {
		return []Request{}, fmt.Errorf("%w request: %w", ErrUnmarshal, err)
	}

	return waiting.Requests, nil
}

func companyNorm(s string) string {
	return strings.SplitN(s, "-В", 2)[0]
}
