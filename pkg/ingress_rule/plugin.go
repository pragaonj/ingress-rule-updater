package ingress_rule

import (
	"context"
	"fmt"
	"github.com/pragaonj/ingress-rule-updater/pkg/ingress_rule/service"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

func RunPlugin(ctx context.Context, configFlags *genericclioptions.ConfigFlags, options *Options) error {
	config, err := configFlags.ToRESTConfig()
	if err != nil {
		return fmt.Errorf("failed to read kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}

	namespace, _, err := configFlags.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return err
	}

	if _, err = clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{}); err != nil {
		return err
	}

	ingressService := service.NewIngressService(clientset, namespace, options.IngressName, options.IngressClassName)

	if options.Set {
		backendRule := service.CreateIngressRule(options.Host, options.Path, options.PathType, options.ServiceName, options.PortNumber)

		created, err := ingressService.AddRule(ctx, backendRule, options.TlsSecret)
		if err != nil {
			return err
		}

		fmt.Printf("Added rule for host '%s' with path '%s' (path type: '%s') for service '%s' (port: '%d') to ingress '%s'\n",
			options.Host, options.Path, options.PathType, options.ServiceName, options.PortNumber, options.IngressName)
		if created {
			fmt.Printf("Created ingress '%s'\n", options.IngressName)
		}
	} else if options.Delete {
		deleted, err := ingressService.DeleteRule(ctx, options.ServiceName, options.PortNumber)
		if err != nil {
			return err
		}

		if options.PortNumber != 0 {
			fmt.Printf("Removed rule(s) for service '%s' (port: '%d') from ingress '%s'\n", options.ServiceName, options.PortNumber, options.IngressName)
		} else {
			fmt.Printf("Removed rule(s) for service '%s' from ingress '%s'\n", options.ServiceName, options.IngressName)
		}
		if deleted {
			fmt.Printf("Deleted ingress '%s'\n", options.IngressName)
		}
	}

	return nil
}
