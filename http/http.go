package http

import (
	"context"

	"github.com/stamm/jetlend/pkg"
)

type HTTP interface {
	ExpectAmount(ctx context.Context, cfg pkg.Config, days int) ([]byte, error)
	// LoansPorfolio(ctx context.Context, cfg pkg.Config) (string, error)
	// Delayed(ctx context.Context, cfg pkg.Config) (int, error)
	// Balance(ctx context.Context, cfg pkg.Config) (int, int, error)
	// PrimaryMarket(ctx context.Context, cfg pkg.Config) (string, error)
	// SecondaryMarket(ctx context.Context, cfg pkg.Config) (string, error)
}
