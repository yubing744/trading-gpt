package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/yubing744/trading-gpt/cmd"
)

func main() {
	now := time.Now()
	fmt.Printf("%s", now.Format("2006-01-02T15:04:05Z07:00"))

	args := os.Args[1:]
	cmd.Execute(context.Background(), args)
}
