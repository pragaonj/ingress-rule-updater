package cli

import (
	"errors"
	"fmt"
	ingress_rule "github.com/pragaonj/ingress-rule-updater/pkg/ingress_rule"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"os"
	"strings"
)

var (
	KubernetesConfigFlags *genericclioptions.ConfigFlags
	cf                    *CliFlags
)

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "ingress-rule",
		Short:         "",
		Long:          `.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				fmt.Println("Usage: <command> <ingress-name>\nWhere command is:\nset\ndelete")
				return errors.New("invalid command line arguments")
			}

			options := CreateOptions(cf, args[0], args[1])
			if options == nil {
				return errors.New("invalid command line flags supplied")
			}

			if err := ingress_rule.RunPlugin(KubernetesConfigFlags, options); err != nil {
				return err
			}

			return nil
		},
	}

	cobra.OnInitialize(initConfig)

	KubernetesConfigFlags = genericclioptions.NewConfigFlags(false)
	KubernetesConfigFlags.AddFlags(cmd.Flags())

	cf = AddOptionFlags(cmd.Flags())

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	return cmd
}

func InitAndExecute() {
	if err := RootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	viper.AutomaticEnv()
}
