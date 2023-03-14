package pkg

import "sync"

type Distr struct {
	Data []Data `json:"data"`
}
type Data struct {
	Debt    float64 `json:"principal_debt"`
	LoadID  uint    `json:"loan_id"`
	Company string  `json:"company"`
}

type Analytics struct {
	Data DataAn `json:"data"`
}
type DataAn struct {
	Status Status `json:"status"`
}
type Status struct {
	Delayed float64 `json:"delayed"`
}
type Details struct {
	Data DataDet `json:"data"`
}
type DataDet struct {
	Balance Balance `json:"balance"`
}
type Balance struct {
	Reserved float64 `json:"reserved"`
	Free     float64 `json:"free"`
}

type Report struct {
	Sum      float64
	Values   map[string]float64
	Delayed  float64
	Reserved float64
	Free     float64
	Mu       sync.Mutex
}

type Expect struct {
	Data []ExpectData `json:"data"`
}
type ExpectData struct {
	Date   int64   `json:"date"`
	Amount float64 `json:"amount"`
}
