package cli

import (
	"errors"
	"github.com/pragaonj/ingress-rule-updater/pkg/ingress_rule"
	"github.com/spf13/cobra"
)

var IngressRuleDeleteConfigFlags *CliFlags

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use: "delete <ingress-name> [flags]",
	Example: "  ingress-rule delete my-ingress --service foo" +
		"\n  ingress-rule delete my-ingress --service foo --port 80",
	Short: "Remove kubernetes ingress rules via command line.",
	Long:  `Deletes a backend rule from an ingress. Deletes the ingress if there are no rules left. Supports removal by service name or a combination of service name and port number.`,
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
		options := CreateOptions(IngressRuleDeleteConfigFlags, COMMAND_DELETE, args[0])
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
	rootCmd.AddCommand(deleteCmd)

	IngressRuleDeleteConfigFlags = AddOptionFlags(deleteCmd.Flags(), COMMAND_DELETE)

	deleteCmd.MarkFlagRequired("service")
}
