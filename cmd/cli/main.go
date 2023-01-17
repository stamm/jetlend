package main

import (
	"context"
	"fmt"
	"os"

	"github.com/stamm/jetlend/pkg"
)

func main() {
	ctx := context.Background()
	msg, err := pkg.Run(ctx, os.Getenv("JETLEND_COOKIE"), true)
	if err != nil {
		panic(err)
		return
	}

	fmt.Print(msg)
}
