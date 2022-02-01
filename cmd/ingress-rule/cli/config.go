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

const COMMAND_SET = "set"
const COMMAND_DELETE = "delete"

func AddOptionFlags(flagSet *pflag.FlagSet, command string) *CliFlags {
	cf := &CliFlags{
		Host:        stringptr(""),
		Path:        stringptr(""),
		PathType:    stringptr(""),
		ServiceName: stringptr(""),
		PortNumber:  intptr(0),
		//todo add support for PortName as alternative to PortNumber
	}

	if command == COMMAND_SET {
		flagSet.StringVar(cf.Host, "host", "", "Set host e.g. foo.example.com, *.example.com, example.com (optional)")
		flagSet.StringVar(cf.Path, "path", "/", "Set matching path (optional)")
		flagSet.StringVar(cf.PathType, "path-type", "prefix", "Set matching type for path (optional); Accepts: \"Prefix\", \"Exact\", \"ImplementationSpecific\"")
	}
	flagSet.StringVar(cf.ServiceName, "service", "", "Name of backend service (must be in the same namespace as the ingress)")
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
	if command != COMMAND_SET && command != COMMAND_DELETE {
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

	path := ""
	pathType := networking.PathTypePrefix

	if command == COMMAND_SET {
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

		pathUri, err := url.ParseRequestURI(*flags.Path)
		if err != nil {
			fmt.Println("Invalid path supplied")
			return nil
		}
		path = pathUri.Path

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
			fmt.Println("Invalid path-type supplied")
			return nil
		}
	}

	return &ingress_rule.Options{
		IngressName: ingressName,
		Host:        *flags.Host,
		Path:        path,
		Delete:      strings.ToLower(command) == COMMAND_DELETE,
		Set:         strings.ToLower(command) == COMMAND_SET,
		PathType:    pathType,
		ServiceName: *flags.ServiceName,
		PortNumber:  int32(*flags.PortNumber),
	}
}
