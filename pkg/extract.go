package pkg

import (
	"encoding/json"
	"fmt"
	"sort"
)

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

func extractDelayed(body []byte) (float64, error) {
	var analyt Analytics
	err := json.Unmarshal(body, &analyt)
	if err != nil {
		return .0, fmt.Errorf("cant unmarshal analitics: %w", err)
	}

	return analyt.Data.Status.Delayed, nil
}

func extractReserved(body []byte) (float64, error) {
	var details Details
	err := json.Unmarshal(body, &details)
	if err != nil {
		return .0, fmt.Errorf("cant unmarshal analitics: %w", err)
	}
	return details.Data.Balance.Reserved, nil

}
