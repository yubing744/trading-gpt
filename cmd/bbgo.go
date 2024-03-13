package cmd

import (
	"context"

	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/cmd"

	log "github.com/sirupsen/logrus"
	_ "github.com/yubing744/trading-gpt/pkg"
)

func init() {
	bbgo.SetWrapperBinary()
}

func Execute(ctx context.Context, args []string) {
	rootCmd := cmd.RootCmd
	rootCmd.SetArgs(args)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.WithError(err).Fatalf("cannot execute command")
	}
}
