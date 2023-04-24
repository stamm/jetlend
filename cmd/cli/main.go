package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/stamm/jetlend/pkg"
)

func main() {
	ctx := context.Background()
	// ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	// defer stop()
	cookies := strings.Split(os.Getenv("JETLEND_COOKIE"), ",")
	d, ok := os.LookupEnv("JETLEND_DAYS")
	if !ok {
		d = "7"
	}
	days, err := strconv.Atoi(d)
	if err != nil {
		panic(err)
	}
	var mode = flag.String("m", "", "stats: empty (stat), expect")
	flag.Parse()

	switch *mode {
	case "", "stat":
		msg, err := pkg.Run(ctx, cookies, true)
		if err != nil {
			panic(err)
		}

		fmt.Print(msg)
	case "expect":
		msg, err := pkg.ExpectAmount(ctx, cookies, true, days)
		if err != nil {
			panic(err)
		}
		fmt.Print(msg)
	}
	// stop()
	// fmt.Printf("ctx done: %s", <-ctx.Done())
}
