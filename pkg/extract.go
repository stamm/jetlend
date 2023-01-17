//go:build go1.20

package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

var ErrUnmarshal = errors.New("can't unmarshal")

func extractLoans(body []byte) (float64, []float64, error) {
	var distr Distr
	err := json.Unmarshal(body, &distr)
	if err != nil {
		return 0., []float64{}, fmt.Errorf("%w loans: %w", ErrUnmarshal, err)
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
