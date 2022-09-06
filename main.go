package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
)

const target = 0.005

type Distr struct {
	Data []Data `json="data"`
}
type Data struct {
	Debt float64 `json:"principal_debt"`
}

func main() {
	body, err := getJson()
	if err != nil {
		panic(err)
	}

	sum, values, err := extractData(body)
	if err != nil {
		panic(err)
	}
	pr(sum, values)
}

func getJson() ([]byte, error) {
	req, err := http.NewRequest("GET", "https://jetlend.ru/invest/api/portfolio/distribution", nil)
	if err != nil {
		return []byte{}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11.15; rv:104.0) Gecko/20100101 Firefox/104.0")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	req.Header.Set("Dnt", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "https://jetlend.ru/invest/v3/analytics")
	req.Header.Set("Cookie", os.Getenv("JETLEND_COOKIE"))
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Te", "trailers")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func extractData(body []byte) (float64, []float64, error) {
	var distr Distr
	err := json.Unmarshal(body, &distr)
	if err != nil {
		return 0., []float64{}, err
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

func pr(sum float64, values []float64) error {
	l := len(values)
	// sum = 1_936_000
	fmt.Printf("sum = %0.f\n", sum)
	fmt.Printf("len = %+v\n", l)
	for _, q := range []int{50, 75, 90, 95, 99, 100} {
		c := int(float64(l) * float64(q) / 100)
		if c >= l {
			c = l - 1
		}
		p := values[c] / sum

		fmt.Printf("%3dq c=%d %5.f %0.2f%%", q, c+1, values[c], p*100.0)
		if p > target {
			fmt.Print("\033[31m")
			fmt.Printf(" > %0.1f%%", target*100)
		}
		fmt.Print("\033[0m\n")
	}
	return nil
}
