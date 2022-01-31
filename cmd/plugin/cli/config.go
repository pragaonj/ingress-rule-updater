package cli

import (
	"fmt"
	"github.com/pragaonj/ingress-rule-updater/pkg/ingress_rule"
	"github.com/spf13/pflag"
	networking "k8s.io/api/networking/v1"
	"net/url"
	"regexp"
	"strings"
)

type CliFlags struct {
	Host        *string
	Path        *string
	PathType    *string
	ServiceName *string
	PortNumber  *int
}

func AddOptionFlags(flagSet *pflag.FlagSet) *CliFlags {
	cf := &CliFlags{
		Host:        stringptr(""),
		Path:        stringptr(""),
		PathType:    stringptr(""),
		ServiceName: stringptr(""),
		PortNumber:  intptr(0),
		//todo add support for PortName as alternative to PortNumber
	}

	flagSet.StringVar(cf.Host, "host", "", "Optional host e.g. foo.example.com, *.example.com, example.com")
	flagSet.StringVar(cf.Path, "path", "/", "Optional path")
	flagSet.StringVar(cf.PathType, "path-type", "prefix", "Path type; possible values: \"Prefix\", \"Exact\", \"ImplementationSpecific\"; defaults to \"prefix\"")
	flagSet.StringVar(cf.ServiceName, "service", "", "Name of the backend service (must be in the same namespace)")
	flagSet.IntVar(cf.PortNumber, "port", 0, "Port number of backend service")

	return cf
}

func stringptr(val string) *string {
	return &val
}
func intptr(val int) *int {
	return &val
}

func CreateOptions(flags *CliFlags, command string, ingressName string) *ingress_rule.Options {
	pathUri, err := url.ParseRequestURI(*flags.Path)
	if err != nil {
		fmt.Println("Invalid path supplied")
		return nil
	}

	var pathType networking.PathType
	switch strings.ToLower(*flags.PathType) {
	case "exact":
		pathType = networking.PathTypeExact
		break
	case "prefix":
		pathType = networking.PathTypePrefix
		break
	case "implementationspecific":
		pathType = networking.PathTypeImplementationSpecific
		break
	default:
		fmt.Println("Invalid path type supplied")
		return nil
		break
	}

	if strings.ToLower(command) != "set" && strings.ToLower(command) != "delete" {
		fmt.Printf("Error: unknown command \"%s\" for \"ingress-rule\"\n", command)
		fmt.Printf("Allowed commands are \"set\" and \"delete\"\n")
		return nil
	}

	if *flags.ServiceName == "" {
		fmt.Println("No service name supplied.")
		return nil
	}

	if ingressName == "" {
		fmt.Printf("Error: no ingress name supplied\n")
		return nil
	}

	if strings.ToLower(command) == "set" {
		if *flags.Host != "" {
			matches, err := regexp.MatchString("^([a-zA-Z0-9-_\\*]+\\.)*[a-zA-Z0-9][a-zA-Z0-9-_]+\\.[a-zA-Z]{2,11}?$", *flags.Host)
			if err != nil {
				fmt.Print(err)
				return nil
			}
			if !matches {
				fmt.Println("Invalid host supplied")
				return nil
			}
		}

		if *flags.PortNumber < 0 || *flags.PortNumber >= 1<<16 {
			fmt.Println("Invalid port supplied")
			return nil
		}

		if *flags.PortNumber == 0 {
			fmt.Println("No port supplied")
			return nil
		}
	}

	return &ingress_rule.Options{
		IngressName: ingressName,
		Host:        *flags.Host,
		Path:        pathUri.Path,
		Delete:      strings.ToLower(command) == "delete",
		Set:         strings.ToLower(command) == "set",
		PathType:    pathType,
		ServiceName: *flags.ServiceName,
		PortNumber:  int32(*flags.PortNumber),
	}
}
