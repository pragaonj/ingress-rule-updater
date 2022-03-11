package main

import (
	"context"
	"github.com/pragaonj/ingress-rule-updater/cmd/ingress-rule/cli"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"os"
	"os/signal"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer ctx.Done()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, os.Kill)

	go func() {
		<-sigs
		cancel()
	}()

	cli.Execute(ctx)
}
