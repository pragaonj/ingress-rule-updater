package service

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

// CreateIngressRule creates a new rule for an ingress.
func CreateIngressRule(hostname string, path string, pathType networking.PathType, backendServiceName string, port int32) *networking.IngressRule {
	return &networking.IngressRule{
		Host: hostname,
		IngressRuleValue: networking.IngressRuleValue{
			HTTP: &networking.HTTPIngressRuleValue{
				Paths: []networking.HTTPIngressPath{{
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

// AddRule configures a new backend rule.
// If an ingress with the given name i.ingressName exists it will be updated, otherwise a new ingress will be created.
// Returns if the ingress has been created and an error
func (i *IngressService) AddRule(ctx context.Context, ingressRule *networking.IngressRule) (created bool, err error) {
	ingress, err := i.kubeIngress.Get(ctx, i.ingressName, metav1.GetOptions{})
	if apperror.IsNotFound(err) {
		// create new ingress if there is no ingress matching the criteria
		return true, i.createIngress(ctx, ingressRule)
	} else if err != nil {
		return false, err
	}

	// check if there is already a rule for this host
	for i1, rule := range ingress.Spec.Rules {
		if rule.Host == ingressRule.Host {
			for _, path := range rule.HTTP.Paths {
				if path.Path == ingressRule.HTTP.Paths[0].Path &&
					*path.PathType == *ingressRule.HTTP.Paths[0].PathType &&
					path.Backend.Service.Name == ingressRule.HTTP.Paths[0].Backend.Service.Name &&
					path.Backend.Service.Port == ingressRule.HTTP.Paths[0].Backend.Service.Port {
					// exact same rule already exists
					return false, ErrorIngressRuleAlreadyExists
				}
			}
			// add rule to for existing host
			ingress.Spec.Rules[i1].HTTP.Paths = append(ingress.Spec.Rules[i1].HTTP.Paths, ingressRule.HTTP.Paths[0])

			// try update and return
			_, err = i.kubeIngress.Update(ctx, ingress, metav1.UpdateOptions{})
			return false, err
		}
	}

	// add new host rule if not exiting
	ingress.Spec.Rules = append(ingress.Spec.Rules, *ingressRule)

	_, err = i.kubeIngress.Update(ctx, ingress, metav1.UpdateOptions{})
	return false, err
}

// DeleteRule removes the rule by service name or service name and port.
// Returns if the resource has been deleted and an error
func (i *IngressService) DeleteRule(ctx context.Context, serviceName string, servicePort int32) (deleted bool, err error) {
	ingress, err := i.kubeIngress.Get(ctx, i.ingressName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	var newRules []networking.IngressRule
	changed := false

	for _, rule := range ingress.Spec.Rules {
		var newPaths []networking.HTTPIngressPath
		for _, p := range rule.HTTP.Paths {
			if !(p.Backend.Service.Name == serviceName && (servicePort == 0 || p.Backend.Service.Port.Number == servicePort)) {
				newPaths = append(newPaths, p)
			} else {
				changed = true
			}
		}
		if len(newPaths) > 0 {
			rule.HTTP.Paths = newPaths
			newRules = append(newRules, rule)
		}
	}

	if len(newRules) == 0 {
		// delete ingress when the last rule is removed
		return true, i.kubeIngress.Delete(ctx, ingress.Name, metav1.DeleteOptions{})
	}

	if changed {
		ingress.Spec.Rules = newRules
		_, err = i.kubeIngress.Update(ctx, ingress, metav1.UpdateOptions{})
		return false, err
	}

	return false, ErrorIngressRuleNotFound
}

var ErrorIngressRuleAlreadyExists = errors.New("ingress rule already exists")
var ErrorIngressRuleNotFound = errors.New("could not find ingress rule for service")
