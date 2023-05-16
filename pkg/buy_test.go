package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuy(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name string
		req  Request
		sum  float64
		exp  float64
	}{
		{
			name: "empty",
			req:  Request{},
			exp:  3_000,
		},
		{
			name: "half",
			req:  Request{},
			sum:  1_500,
			exp:  1_500,
		},
		{
			name: "oversum",
			req:  Request{},
			sum:  3_000,
			exp:  0,
		},
		{
			name: "collected",
			req:  Request{CollectedPercentage: 100},
			exp:  0.,
		},

		{
			name: ">30%",
			req:  Request{InterestRate: 0.3},
			exp:  1_000,
		},
		{
			name: ">30% half",
			req:  Request{InterestRate: 0.3},
			sum:  500,
			exp:  500,
		},
		{
			name: ">30% enoght",
			req:  Request{InterestRate: 0.3},
			sum:  1_000,
			exp:  0,
		},

		{
			name: ">=1y",
			req:  Request{Term: 390},
			sum:  0,
			exp:  1_500,
		},
		{
			name: ">=1y half",
			req:  Request{Term: 390},
			sum:  750,
			exp:  750,
		},
		{
			name: ">=1y enoght",
			req:  Request{Term: 390},
			sum:  1_500,
			exp:  0,
		},
	}

	max := 6_000.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buy := muchBuy(tc.req, max, tc.sum)
			assert.Equal(tc.exp, buy)
		})
	}
}
