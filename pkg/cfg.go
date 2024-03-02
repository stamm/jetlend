package pkg

import "flag"

type Cfg struct {
	MinTargetPercent float64 `json:"min_target_percent"`
	MinTargetSum     float64 `json:"min_target_sum"`
	MaxTargetPercent float64 `json:"max_target_percent"`
	MaxTargetSum     float64 `json:"max_target_sum"`
}

func NewCfg() Cfg {
	cfg := DefaultCfg()
	var minTargetPercent = flag.Float64("min-target-percent", -1, "min target: 0.1 - 10%")
	var minTargetSum = flag.Float64("min-target-sum", -1, "min target sum")
	var maxTargetPercent = flag.Float64("max-target-percent", -1, "max target: 0.002 - 0.2%")
	var maxTargetSum = flag.Float64("max-target-sum", -1, "max target sum")
	flag.Parse()
	if *minTargetPercent != -1 {
		cfg.MinTargetPercent = *minTargetPercent
	}
	if *minTargetSum != -1 {
		cfg.MinTargetSum = *minTargetSum
	}
	if *maxTargetPercent != -1 {
		cfg.MaxTargetPercent = *maxTargetPercent
	}
	if *maxTargetSum != -1 {
		cfg.MaxTargetSum = *maxTargetSum
	}
	return cfg
}
func DefaultCfg() Cfg {
	return Cfg{
		MinTargetPercent: 0.00206,
		MinTargetSum:     6001.,
		MaxTargetPercent: 0.003,
		MaxTargetSum:     9001.,
	}
}
