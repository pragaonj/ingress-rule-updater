package ingress_rule

import (
	"context"
	"errors"
	"fmt"
	"github.com/pragaonj/ingress-rule-updater/pkg/ingress_rule/service"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
)

func RunPlugin(configFlags *genericclioptions.ConfigFlags, options *Options) error {
	config, err := configFlags.ToRESTConfig()
	if err != nil {
		return fmt.Errorf("failed to read kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}

	ctx := context.TODO()

	//fmt.Println("Context: " + *configFlags.Context)
	//fmt.Println("Namespace: " + *configFlags.Namespace)

	namespace := *configFlags.Namespace
	if namespace == "" {
		namespace = "default"
	}

	exists, err := NamespaceExists(ctx, clientset, namespace)
	if err != nil {
		return err
	}
	if !exists {
		fmt.Println("Cloud not find namespace")
		return errors.New("invalid command line flags supplied")
	}
	fmt.Printf("Using namespace: %s\n", namespace)

	ingressService := service.NewIngressService(clientset, namespace, options.IngressName)

	if options.Set {
		backendRule := service.CreateIngressRule(options.Host, options.Path, options.PathType, options.ServiceName, options.PortNumber)

		err := ingressService.AddRule(ctx, backendRule)
		if err != nil {
			return err
		}
		fmt.Printf("Added rule for backend service: %s\n", options.ServiceName)
	} else if options.Delete {
		err := ingressService.DeleteRule(ctx, options.ServiceName, options.PortNumber)
		if err != nil {
			return err
		}
		fmt.Printf("Removed rule for backend service: %s\n", options.ServiceName)
	}

	return nil
}

func NamespaceExists(ctx context.Context, clientset *kubernetes.Clientset, namespace string) (bool, error) {
	_, err := clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	return true, nil
}
