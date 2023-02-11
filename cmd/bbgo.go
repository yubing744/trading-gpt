package main

import (
	"fmt"
	"time"

	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/cmd"
	_ "github.com/yubing744/trading-bot"
)

func init() {
	bbgo.SetWrapperBinary()
}

func main() {
	now := time.Now()
	fmt.Printf("%s", now.Format("2006-01-02T15:04:05Z07:00"))
	cmd.Execute()
}
