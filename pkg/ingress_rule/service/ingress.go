package service

import (
	"context"
	"errors"
	networking "k8s.io/api/networking/v1"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	clientnetworking "k8s.io/client-go/kubernetes/typed/networking/v1"
	"log"
)

type IngressService struct {
	kubeIngress      clientnetworking.IngressInterface
	ingressName      string
	ingressClassName string
}

func NewIngressService(clientset *kubernetes.Clientset, namespace string, ingressName string, ingressClassName string) *IngressService {
	return &IngressService{
		kubeIngress:      clientset.NetworkingV1().Ingresses(namespace),
		ingressName:      ingressName,
		ingressClassName: ingressClassName,
	}
}

func (i *IngressService) createIngress(ctx context.Context, ingressRule *networking.IngressRule, tlsSecret string) error {
	ingressClass := &i.ingressClassName
	if *ingressClass == "" {
		ingressClass = nil
	}

	ingress := &networking.Ingress{
		TypeMeta: meta.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: i.ingressName,
		},
		Spec: networking.IngressSpec{
			IngressClassName: ingressClass,
			Rules:            []networking.IngressRule{*ingressRule},
		},
	}

	if tlsSecret != "" && ingressRule.Host != "" {
		ingress.Spec.TLS = []networking.IngressTLS{{
			Hosts:      []string{ingressRule.Host},
			SecretName: tlsSecret,
		}}
	}

	_, err := i.kubeIngress.Create(ctx, ingress, meta.CreateOptions{})
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
func (i *IngressService) AddRule(ctx context.Context, ingressRule *networking.IngressRule, tlsSecret string) (created bool, err error) {
	ingress, err := i.kubeIngress.Get(ctx, i.ingressName, meta.GetOptions{})
	if apierror.IsNotFound(err) {
		// create new ingress if there is no ingress matching the criteria
		return true, i.createIngress(ctx, ingressRule, tlsSecret)
	} else if err != nil {
		return false, err
	}

	// check if there is already a rule for this host (and add the path)
	exists, err := addPathToExistingHostIfRuleExists(ingress, ingressRule)
	if err != nil {
		return false, err
	}
	if !exists {
		// add new host rule if not existing
		ingress.Spec.Rules = append(ingress.Spec.Rules, *ingressRule)
	}

	err = addTlsRuleIfSecretIsSupplied(ingress, ingressRule.Host, tlsSecret)
	if err != nil {
		return false, err
	}

	_, err = i.kubeIngress.Update(ctx, ingress, meta.UpdateOptions{})
	return false, err
}

// addPathToExistingHostIfRuleExists checks if the ingress already contains a rule for the given host. If so, the function trys to add a new path to this rule.
// Will return true if the rule has been added, will throw an ErrorIngressRuleAlreadyExists error if the same rule (same host and path) already exists.
func addPathToExistingHostIfRuleExists(ingress *networking.Ingress, ingressRule *networking.IngressRule) (bool, error) {
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
			return true, nil
		}
	}

	return false, nil
}

// DeleteRule removes the rule by service name or service name and port.
// Returns if the resource has been deleted and an error
func (i *IngressService) DeleteRule(ctx context.Context, serviceName string, servicePort int32) (deleted bool, err error) {
	ingress, err := i.kubeIngress.Get(ctx, i.ingressName, meta.GetOptions{})
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
		return true, i.kubeIngress.Delete(ctx, ingress.Name, meta.DeleteOptions{})
	}

	if changed {
		ingress.Spec.Rules = newRules
		deleteTlsRulesForNoLongerExistingHosts(ingress)
		_, err = i.kubeIngress.Update(ctx, ingress, meta.UpdateOptions{})
		return false, err
	}

	return false, ErrorIngressRuleNotFound
}

func addTlsRuleIfSecretIsSupplied(ingress *networking.Ingress, host string, tlsSecret string) error {
	if tlsSecret == "" || host == "" {
		return nil
	}

	secretIndex := -1

	for i, tlsEntry := range ingress.Spec.TLS {
		for _, existingHost := range tlsEntry.Hosts {
			if host == existingHost {
				if tlsEntry.SecretName != tlsSecret {
					// do nothing and return error to user
					return ErrorTlsConfigurationAlreadyExists
				} else {
					// do nothing since correct tls configuration already exists
					return nil
				}
			}
		}
		if tlsEntry.SecretName == tlsSecret {
			secretIndex = i
		}
	}

	if secretIndex == -1 {
		ingress.Spec.TLS = append(ingress.Spec.TLS, networking.IngressTLS{
			Hosts:      []string{host},
			SecretName: tlsSecret,
		})
	} else {
		ingress.Spec.TLS[secretIndex].Hosts = append(ingress.Spec.TLS[secretIndex].Hosts, host)
	}

	return nil
}

func deleteTlsRulesForNoLongerExistingHosts(ingress *networking.Ingress) {
	var newTlsConfig []networking.IngressTLS

	for i, tlsEntry := range ingress.Spec.TLS {
		for i2, host := range tlsEntry.Hosts {
			hostExists := false
			for _, rule := range ingress.Spec.Rules {
				if rule.Host == host {
					hostExists = true
				}
			}
			if !hostExists {
				ingress.Spec.TLS[i].Hosts[i2] = ingress.Spec.TLS[i].Hosts[len(ingress.Spec.TLS[i].Hosts)-1]
				ingress.Spec.TLS[i].Hosts = ingress.Spec.TLS[i].Hosts[:len(ingress.Spec.TLS[i].Hosts)-1]
				break
			}
		}
		log.Printf("%+v\n", ingress.Spec.TLS[i].Hosts)
		if len(ingress.Spec.TLS[i].Hosts) != 0 {
			newTlsConfig = append(newTlsConfig, ingress.Spec.TLS[i])
		}
	}

	ingress.Spec.TLS = newTlsConfig
}

var ErrorIngressRuleAlreadyExists = errors.New("ingress rule already exists")
var ErrorIngressRuleNotFound = errors.New("could not find ingress rule for service")
var ErrorTlsConfigurationAlreadyExists = errors.New("tls configuration for hostname already exists")
