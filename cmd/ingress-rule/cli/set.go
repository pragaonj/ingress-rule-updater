package cli

import (
	"errors"
	"github.com/pragaonj/ingress-rule-updater/pkg/ingress_rule"
	"github.com/spf13/cobra"
)

var IngressRuleSetConfigFlags *CliFlags

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use: "set <ingress-name> [flags]",
	Example: "  kubectl ingress-rule set my-ingress --service foo --port 80 --host *.foo.com" +
		"\n  kubectl ingress-rule set my-ingress --service foo --port 80 --host example.com --path /foo",
	Short: "Add kubernetes ingress rules via command line. If the ingress does not exist a new ingress will be created.",
	Long:  `Adds a backend rule to an ingress. If the ingress does not exist a new ingress will be created.`,
	Args: func(cmd *cobra.Command, args []string) error {
		// validate ingress-name arg
		if len(args) < 1 {
			return errors.New("no ingress name was specified")
		} else if len(args) > 1 {
			return errors.New("invalid number of command line arguments; only a ingress name is expected")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		options := CreateOptions(IngressRuleSetConfigFlags, COMMAND_SET, args[0])
		if options == nil {
			return errors.New("invalid command line flags supplied")
		}

		if err := ingress_rule.RunPlugin(cmd.Context(), KubernetesConfigFlags, options); err != nil {
			return err
		}

		return nil
	},
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	rootCmd.AddCommand(setCmd)

	IngressRuleSetConfigFlags = AddOptionFlags(setCmd.Flags(), COMMAND_SET)

	setCmd.MarkFlagRequired("service")
	setCmd.MarkFlagRequired("port")
}
