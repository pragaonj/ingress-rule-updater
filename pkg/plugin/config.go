package plugin

import (
	"fmt"
	"github.com/spf13/pflag"
	networking "k8s.io/api/networking/v1"
	"net/url"
	"regexp"
	"strings"
)

type CliFlags struct {
	IngressName *string
	Host        *string
	Path        *string
	Add         *bool
	Delete      *bool
	Update      *bool
	PathType    *string
	ServiceName *string
	PortNumber  *int
}

type Options struct {
	IngressName string
	Host        string
	Path        string
	Add         bool
	Delete      bool
	Update      bool
	PathType    networking.PathType
	ServiceName string
	PortNumber  int32
}

func AddOptionFlags(flagSet *pflag.FlagSet) *CliFlags {
	cf := &CliFlags{
		IngressName: stringptr(""),
		Host:        stringptr(""),
		Path:        stringptr(""),
		Add:         boolptr(false),
		Delete:      boolptr(false),
		Update:      boolptr(false),
		PathType:    stringptr(""),
		ServiceName: stringptr(""),
		PortNumber:  intptr(0),
		//todo add support for PortName as alternative to PortNumber
	}

	flagSet.StringVar(cf.IngressName, "ingress-name", "", "Ingress name")
	flagSet.StringVar(cf.Host, "host", "", "Optional host e.g. foo.example.com, *.example.com, example.com")
	flagSet.StringVar(cf.Path, "path", "/", "Optional path")
	flagSet.StringVar(cf.PathType, "path-type", "prefix", "Path type; possible values: \"Prefix\", \"Exact\", \"ImplementationSpecific\"; defaults to \"prefix\"")
	flagSet.StringVar(cf.ServiceName, "service-name", "", "Name of the backend service (must be in the same namespace)")
	flagSet.IntVar(cf.PortNumber, "port", 0, "Port number of backend service")
	flagSet.BoolVar(cf.Add, "add", false, "Add new rule")
	flagSet.BoolVar(cf.Update, "update", false, "Update existing rule")
	flagSet.BoolVar(cf.Delete, "delete", false, "Delete existing rule")

	return cf
}

func stringptr(val string) *string {
	return &val
}
func intptr(val int) *int {
	return &val
}
func boolptr(val bool) *bool {
	return &val
}

func CreateOptions(flags *CliFlags) *Options {
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

	if (*flags.Add || *flags.Update) && *flags.Delete {
		fmt.Println("Invalid combination of operations supplied. delete cannot be used in combination with update or add")
		return nil
	}

	if !*flags.Add && !*flags.Update && !*flags.Delete {
		fmt.Println("No operation supplied.")
		return nil
	}

	if *flags.ServiceName == "" {
		fmt.Println("No service name supplied.")
		return nil
	}

	if *flags.IngressName == "" {
		fmt.Println("No ingress name supplied.")
		return nil
	}

	if !*flags.Delete {
		matches, err := regexp.MatchString("^([a-zA-Z0-9-_\\*]+\\.)*[a-zA-Z0-9][a-zA-Z0-9-_]+\\.[a-zA-Z]{2,11}?$", *flags.Host)
		if err != nil {
			fmt.Print(err)
			return nil
		}
		if !matches {
			fmt.Println("Invalid host supplied")
			return nil
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

	return &Options{
		IngressName: *flags.IngressName,
		Host:        *flags.Host,
		Path:        pathUri.Path,
		Add:         *flags.Add,
		Delete:      *flags.Delete,
		Update:      *flags.Update,
		PathType:    pathType,
		ServiceName: *flags.ServiceName,
		PortNumber:  int32(*flags.PortNumber),
	}
}
