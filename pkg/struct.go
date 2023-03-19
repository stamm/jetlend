package pkg

import (
	"encoding/json"
	"errors"
	"sync"
)

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
	Requests []Request
	Mu       sync.Mutex
}

func (r *Report) WithMutex(f func(r *Report)) {
	r.Mu.Lock()
	f(r)
	r.Mu.Unlock()
}

type Expect struct {
	Data []ExpectData `json:"data"`
}
type ExpectData struct {
	Date   int64   `json:"date"`
	Amount float64 `json:"amount"`
}

type Waiting struct {
	Requests []Request `json:"requests"`
}
type Request struct {
	Amount              CustomFloat64 `json:"amount"`
	InterestRate        float64       `json:"interest_rate"`
	CollectedPercentage float64       `json:"collected_percentage"`
	InvestingAmount     CustomFloat64 `json:"investing_amount"`
	Rating              string        `json:"rating"`
	LoanName            string        `json:"loan_name"`
}

type CustomFloat64 struct {
	Float64 float64
}

func (cf *CustomFloat64) UnmarshalJSON(data []byte) error {
	if data[0] == 34 {
		err := json.Unmarshal(data[1:len(data)-1], &cf.Float64)
		if err != nil {
			return errors.New("CustomFloat64: UnmarshalJSON: " + err.Error())
		}
	} else {
		err := json.Unmarshal(data, &cf.Float64)
		if err != nil {
			return errors.New("CustomFloat64: UnmarshalJSON: " + err.Error())
		}
	}
	return nil
}

func (cf CustomFloat64) MarshalJSON() ([]byte, error) {
	json, err := json.Marshal(cf.Float64)
	return json, err
}
