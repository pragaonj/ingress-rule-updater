package cli

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"os"
	"strings"
)

var (
	KubernetesConfigFlags *genericclioptions.ConfigFlags
)

var version = "dev"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "ingress-rule <command> <ingress-name> [flags]",
	Example: "  kubectl ingress-rule set my-ingress --service foo --port 80 --host *.foo.com" +
		"\n  kubectl ingress-rule set my-ingress --service foo --port 80 --host example.com --path /foo" +
		"\n\n  kubectl ingress-rule delete my-ingress --service foo" +
		"\n  kubectl ingress-rule delete my-ingress --service foo --port 80",
	Short:   "Add/remove kubernetes ingress rules via command line.",
	Long:    `Add/remove kubernetes ingress rules via command line.`,
	Version: version,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		viper.BindPFlags(cmd.Flags())
	},
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ctx context.Context) {
	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	KubernetesConfigFlags = genericclioptions.NewConfigFlags(false)
	KubernetesConfigFlags.AddFlags(rootCmd.PersistentFlags())

	cobra.OnInitialize(initConfig)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	//disable completion for now
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func initConfig() {
	viper.AutomaticEnv()
}
