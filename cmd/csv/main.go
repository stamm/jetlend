package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gocarina/gocsv"
)

type Transaction struct {
	Date    string  `csv:"Дата"`
	Type    string  `csv:"Тип операции"`
	Company string  `csv:"Компания"`
	Inn     string  `csv:"ИНН"`
	In      float64 `csv:"Приход"`
	Out     float64 `csv:"Расход"`
	Debt    float64 `csv:"Основной долг"`
	Income  float64 `csv:"Доход"`
}

type FloatCSV float64

func (f *FloatCSV) MarshalCSV() (string, error) {
	return fmt.Sprintf("%.2f", float64(*f)), nil
}

// TYPE;DATE;TICKER;QUANTITY;PRICE;FEE;NKD;NOMINAL;CURRENCY;FEE_CURRENCY;NOTE;LINK_ID;TRADE_SYSTEM_ID
type IntelInvest struct {
	Type          string   `csv:"TYPE"`
	Date          string   `csv:"DATE"`
	Ticker        string   `csv:"DATE"`
	Quantity      string   `csv:"QUANTITY"`
	Price         FloatCSV `csv:"PRICE"`
	Fee           string   `csv:"FEE"`
	Nkd           string   `csv:"NKD"`
	Nominal       string   `csv:"NOMINAL"`
	Currency      string   `csv:"CURRENCY"`
	FeeCurrency   string   `csv:"FEE_CURRENCY"`
	Note          string   `csv:"NOTE"`
	LinkId        string   `csv:"LINK_ID"`
	TradeSystemId string   `csv:"TRADE_SYSTEM_ID"`
}

// type IntelInvest struct {
// 	Type          string
// 	Date          string
// 	Ticker        string
// 	Quantity      string
// 	Price         FloatCSV
// 	Fee           string
// 	Nkd           string
// 	Nominal       string
// 	Currency      string
// 	FeeCurrency   string
// 	Note          string
// 	LinkId        string
// 	TradeSystemId string
// }

type Default struct {
	Company string
	INN     string
	Date    string
}

func main() {
	in, err := os.Open("transactions-2.csv")
	if err != nil {
		panic(err)
	}
	defer in.Close()

	transactions := []*Transaction{}

	if err := gocsv.UnmarshalFile(in, &transactions); err != nil {
		panic(err)
	}
	// for _, v := range []int{0, 11, 63, 122, 953, 3031, 4416, 7182, 7429} {
	// 	fmt.Printf("%+v\n", transactions[v])
	// }

	// for _, client := range clients {
	// 	if client.Type == "investment" ||
	// 		client.Type == "collection" ||
	// 		client.Type == "sale" ||
	// 		client.Type == "default" ||
	// 		client.Type == "contract" ||
	// 		client.Type == "purchase" ||
	// 		client.Type == "payment" {
	// 		continue
	// 	}
	// 	fmt.Printf("=%+v\n", client)
	// 	break
	// }
	sum := FloatCSV(0)
	date := ""
	defaults := make(map[string]Default)
	converted := make([]IntelInvest, 0, 10)
	for i := len(transactions) - 1; i >= 0; i-- {
		transaction := transactions[i]
		// if transaction.Type != "investment" {
		// 	continue
		// }
		if transaction.Type == "default" {
			defaults[transaction.Inn] = Default{
				Company: transaction.Company,
				Date:    transaction.Date,
				INN:     transaction.Inn,
			}
		}
		if v, ok := convert(*transaction); ok {
			if v.Type == "INCOME" && v.Note != "Бонусы" {
				if v.Date != date || i == 0 {
					vSum := v
					if date != "" && sum != 0 {
						vSum.Price = FloatCSV(sum)
					}
					converted = append(converted, vSum)
					date = v.Date
					sum = FloatCSV(0)
				}
				sum += v.Price
			} else {
				converted = append(converted, v)
			}
		}
		// fmt.Printf("=%+v\n", transaction)
	}
	// fmt.Printf("converted: %+v\n", converted)
	// fmt.Printf("defaults: %+v\n", defaults)

	defSum := make(map[string]float64, len(defaults))

	for _, v := range transactions {
		if _, ok := defaults[v.Inn]; !ok {
			continue
		}
		// 		+ contract, purchase
		// - sale, payment

		switch v.Type {
		case "contract":
			defSum[v.Inn] += v.Out
		case "purchase":
			defSum[v.Inn] += v.Debt
		case "payment":
			defSum[v.Inn] -= v.Debt
		case "sale":
			defSum[v.Inn] -= v.Debt
		}
	}
	// fmt.Printf("defSum: %+v\n", defSum)
	for _, v := range defaults {
		// fmt.Printf("inn: %s, company: %s, date %s, sum: %0.2f\n", k, v.Company, v.Date, defSum[v.INN])
		converted = append(converted, MakeII("LOSS", "Дефолт "+v.Company, convertDate(v.Date), defSum[v.INN]))
	}

	gocsv.SetCSVWriter(func(out io.Writer) *gocsv.SafeCSVWriter {
		csvout := gocsv.NewSafeCSVWriter(csv.NewWriter(out))
		csvout.Comma = ';'
		return csvout
	})
	csv, err := gocsv.MarshalStringWithoutHeaders(&converted)
	if err != nil {
		panic(err)
	}
	csv = "#CsvFormatVersion:v1\n" + csv
	fmt.Println(csv)
}

func convert(a Transaction) (IntelInvest, bool) {
	date := convertDate(a.Date)
	switch a.Type {
	// 12362 payment
	// 1219 contract
	// 715 purchase
	// 166 investment
	//  77 sale
	//  46 default
	//  26 collection
	// вводы, выводы, бонусы
	case "investment":
		if a.Income >= 0.01 {
			// fmt.Printf("%+v\n", a)
			return MakeII("INCOME", "Бонусы", date, a.Income), true
		}
		if a.In > 0 {
			return MakeII("MONEYDEPOSIT", "", date, a.In), true
		} else {
			return MakeII("MONEYWITHDRAW", "", date, -a.Out), true
		}
	// прибыль
	case "payment":
		if a.Income > 0 {
			return MakeII("INCOME", "Прибыль", date, a.Income), true
		}
	// продажа
	case "sale":
		if a.Income > 0 {
			return MakeII("INCOME", "Продажа займа", date, a.Income), true
		}
		if a.Income < 0 {
			return MakeII("LOSS", "Продажа займа", date, a.Income), true
		}
	// покупка
	case "purchase":
		if a.Income > 0 {
			return MakeII("INCOME", "Покупка займа", date, a.Income), true
		}
		if a.Income < 0 {
			return MakeII("LOSS", "Покупка займа", date, a.Income), true
		}
		// same
		// 5277,2022-12-08,payment,ИП Шуляка Александр Андреевич,910901600445,27.5,,27.5,0
		// 5279,2022-12-08,collection,ИП Шуляка Александр Андреевич,910901600445,27.5,,27.5,0
		// case "collection":
		// 	return MakeII("INCOME", "Возврат долга", date, a.In), true
	}
	return IntelInvest{}, false
}

func MakeII(t, note, date string, price float64) IntelInvest {
	return IntelInvest{
		Type:     t,
		Date:     date,
		Price:    FloatCSV(price),
		Currency: "RUB",
		Note:     note,
	}

}

func convertDate(in string) string {
	parseTime, err := time.Parse("2006-01-02", in)
	if err != nil {
		panic(err)
	}
	return parseTime.Format("02.01.2006")
}

/*
invest: in - внёс, out - вывел, income - бонус
contract: out - покупка займа
payment: in = долг + income
collection: debt
default: пустой

+ contract, purchase
- sale

*/
