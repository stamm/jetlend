package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractLoans(t *testing.T) {
	assert := assert.New(t)
	d := []byte(`
{"status":"OK","data":[{"company":"ПЛАТОН-В01","principal_debt":1347.18,"loan_id":1450,"amount":5226,"status":"delayed"},{"company":"БеляевВА-В04","principal_debt":4100.0,"loan_id":12079,"amount":4100,"status":"active"}]}
`)
	sum, values, err := extractLoans(d)
	assert.Nil(err)
	assert.Equal(5447.18, sum)
	assert.Equal([]float64{1347.18, 4100.0}, values)
}

func TestExtractLoansWrong(t *testing.T) {
	assert := assert.New(t)
	d := []byte(`
{"status":"OK"
`)
	_, _, err := extractLoans(d)
	assert.ErrorIs(err, ErrUnmarshal)
	assert.ErrorContains(err, "unexpected end of JSON input")
}

func TestExtractDelayed(t *testing.T) {
	assert := assert.New(t)
	d := []byte(`
{"status":"OK","data":{"status":{"delayed":2.34}}}

`)
	delayed, err := extractDelayed(d)
	assert.Nil(err)
	assert.Equal(2.34, delayed)
}

func TestExtractDelayedWrong(t *testing.T) {
	assert := assert.New(t)
	d := []byte(`
{"status":"OK"

`)
	delayed, err := extractDelayed(d)
	assert.ErrorIs(err, ErrUnmarshal)
	assert.ErrorContains(err, "unexpected end of JSON input")
	assert.Equal(.0, delayed)
}

func TestExtractDelayedWithoutStatus(t *testing.T) {
	assert := assert.New(t)
	d := []byte(`
{"status":"OK"}

`)
	delayed, err := extractDelayed(d)
	assert.Nil(err)
	assert.Equal(.0, delayed)
}
func TestExtractBalance(t *testing.T) {
	assert := assert.New(t)
	d := []byte(`
{"status":"OK","data":{"balance":{"reserved":1.23, "free": 4.56}}}
`)
	reserved, free, err := extractBalance(d)
	assert.Nil(err)
	assert.Equal(1.23, reserved)
	assert.Equal(4.56, free)
}

func TestExtractBalanceWrong(t *testing.T) {
	assert := assert.New(t)
	d := []byte(`
{"status":"OK"

`)
	reserved, free, err := extractBalance(d)
	assert.ErrorIs(err, ErrUnmarshal)
	assert.ErrorContains(err, "unexpected end of JSON input")
	assert.Equal(.0, reserved)
	assert.Equal(.0, free)
}

func TestExtractBalanceWithoutStatus(t *testing.T) {
	assert := assert.New(t)
	d := []byte(`
{"status":"OK"}

`)
	reserved, free, err := extractBalance(d)
	assert.Nil(err)
	assert.Equal(.0, reserved)
	assert.Equal(.0, free)
}
