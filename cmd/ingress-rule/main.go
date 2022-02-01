package main

import (
	"context"
	"github.com/pragaonj/ingress-rule-updater/cmd/ingress-rule/cli"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx := context.Background()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)

	go func() {
		<-sigs
		ctx.Done()
	}()

	cli.Execute(ctx)
}
