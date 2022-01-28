package plugin

import (
	"context"
	"errors"
	networking "k8s.io/api/networking/v1"
	apperror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	networkingtypes "k8s.io/client-go/kubernetes/typed/networking/v1"
)

type IngressService struct {
	kubeIngress networkingtypes.IngressInterface
	ingressName string
}

func NewIngressService(clientset *kubernetes.Clientset, namespace string, ingressname string) *IngressService {
	return &IngressService{
		kubeIngress: clientset.NetworkingV1().Ingresses(namespace),
		ingressName: ingressname,
	}
}

func (i *IngressService) createIngress(ctx context.Context, ingressRule *networking.IngressRule) error {
	ingress := &networking.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: i.ingressName,
		},
		Spec: networking.IngressSpec{
			Rules: []networking.IngressRule{*ingressRule},
		},
	}

	_, err := i.kubeIngress.Create(ctx, ingress, metav1.CreateOptions{})

	return err
}

func (i *IngressService) deleteIngress(ctx context.Context, ingressname string) error {
	err := i.kubeIngress.Delete(ctx, ingressname, metav1.DeleteOptions{})

	return err
}

// todo add functionality to add a path to a rule

// CreateIngressRule creates a new rule for an ingress.
func CreateIngressRule(hostname string, path string, pathType networking.PathType, backendServiceName string, port int32) *networking.IngressRule {
	return &networking.IngressRule{
		Host: hostname,
		IngressRuleValue: networking.IngressRuleValue{
			HTTP: &networking.HTTPIngressRuleValue{
				Paths: []networking.HTTPIngressPath{{ //todo add functionality to add a path to a rule right here
					Path:     path,
					PathType: &pathType,
					Backend: networking.IngressBackend{
						Service: &networking.IngressServiceBackend{
							Name: backendServiceName,
							Port: networking.ServiceBackendPort{
								Number: port, //todo add optional port configuration via PortName right here
							},
						},
					},
				}},
			},
		},
	}
}

func (i *IngressService) AddRule(ctx context.Context, ingressRule *networking.IngressRule) error {
	if ingressRule.Host == "" {
		return IngressSpecWithoutHostError
	}
	ingress, err := i.kubeIngress.Get(ctx, i.ingressName, metav1.GetOptions{})
	if apperror.IsNotFound(err) {
		//create new ingress if there is no ingress matching the criteria
		return i.createIngress(ctx, ingressRule)
	} else if err != nil {
		return err
	}

	for _, rule := range ingress.Spec.Rules {
		if rule.Host == ingressRule.Host {
			return IngressRuleForHostAlreadyExists
		}
	}

	ingress.Spec.Rules = append(ingress.Spec.Rules, *ingressRule)

	_, err = i.kubeIngress.Update(ctx, ingress, metav1.UpdateOptions{})
	return err
}

//RemoveRule removes the rule by service name
func (i *IngressService) RemoveRule(ctx context.Context, name string) error {
	ingress, err := i.kubeIngress.Get(ctx, i.ingressName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	for index, rule := range ingress.Spec.Rules {
		for _, path := range rule.HTTP.Paths {
			if path.Backend.Service.Name == name {
				if len(ingress.Spec.Rules) == 1 {
					//delete ingress when the last rule is removed
					return i.deleteIngress(ctx, ingress.Name)
				}

				ingress.Spec.Rules = append(ingress.Spec.Rules[0:index], ingress.Spec.Rules[index+1:]...)

				_, err = i.kubeIngress.Update(ctx, ingress, metav1.UpdateOptions{})
				return err
			}
		}
	}

	return IngressRuleNotFound
}

var IngressSpecWithoutHostError = errors.New("ingress spec without host is not supported at the moment")
var IngressRuleForHostAlreadyExists = errors.New("ingress rule for host already exists")
var IngressRuleNotFound = errors.New("could not find ingress rule for host")
