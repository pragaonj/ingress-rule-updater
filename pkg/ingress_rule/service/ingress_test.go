package service

import (
	"context"
	"github.com/stretchr/testify/assert"
	networking "k8s.io/api/networking/v1"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/typed/networking/v1/fake"
	clienttesting "k8s.io/client-go/testing"
	"testing"
)

func TestIngressService_AddRuleToNewIngress(t *testing.T) {
	tests := []struct {
		name                     string
		newRule                  networking.IngressRule
		ingressCreated           bool
		err                      error
		ingressClassName         string
		tlsSecret                string
		expectedTlsConfiguration []networking.IngressTLS
	}{
		{
			"add rule to new ingress",
			ruleHostFoo(),
			true,
			nil,
			"ingress-class",
			"",
			nil,
		},
		{
			"add rule to new ingress with custom ingress class name",
			ruleHostFoo(),
			true,
			nil,
			"my-ingress-class",
			"",
			nil,
		},
		{
			"add rule to new ingress with tls secret",
			ruleHostFoo(),
			true,
			nil,
			"my-ingress-class",
			"my-secret",
			[]networking.IngressTLS{{
				Hosts:      []string{ruleHostFoo().Host},
				SecretName: "my-secret",
			}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var ingress *networking.Ingress

			f := clienttesting.Fake{}
			f.AddReactor("get", "ingresses", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, errors2.NewNotFound(action.GetResource().GroupResource(), action.(clienttesting.GetAction).GetName())
			})
			f.AddReactor("create", "ingresses", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
				spec := action.(clienttesting.UpdateAction).GetObject().(*networking.Ingress).Spec
				assert.Len(t, spec.Rules, 1)
				assert.Equal(t, test.newRule, spec.Rules[0])
				assert.Equal(t, test.expectedTlsConfiguration, spec.TLS)

				ingress = action.(clienttesting.CreateAction).GetObject().(*networking.Ingress)
				return true, action.(clienttesting.CreateAction).GetObject(), nil
			})

			ingressService := IngressService{
				kubeIngress:      &fake.FakeIngresses{Fake: &fake.FakeNetworkingV1{Fake: &f}},
				ingressName:      "foo",
				ingressClassName: test.ingressClassName,
			}

			created, err := ingressService.AddRule(context.TODO(), &test.newRule, test.tlsSecret)
			assert.Equal(t, test.err, err)
			assert.Equal(t, test.ingressCreated, created)

			assert.Equal(t, test.newRule, ingress.Spec.Rules[0])
			assert.Equal(t, test.ingressClassName, *ingress.Spec.IngressClassName)
			assert.Equal(t, test.expectedTlsConfiguration, ingress.Spec.TLS)
		})
	}
}

func TestIngressService_AddRuleToExistingIngress(t *testing.T) {
	tests := []struct {
		name                     string
		existingRules            []networking.IngressRule
		newRule                  networking.IngressRule
		expectedRules            []networking.IngressRule
		err                      error
		tlsSecret                string
		expectedTlsConfiguration []networking.IngressTLS
		initialTlsConfiguration  []networking.IngressTLS
	}{
		{
			name:          "add rule to exiting ingress",
			existingRules: []networking.IngressRule{ruleHostBar()},
			newRule:       ruleHostFoo(),
			expectedRules: []networking.IngressRule{ruleHostBar(), ruleHostFoo()},
		},
		{
			name:          "add rule to existing ingress with existing host",
			existingRules: []networking.IngressRule{ruleHostFoo()},
			newRule:       ruleHostFoo2(),
			expectedRules: []networking.IngressRule{ruleHostFooTwoRules()},
		},
		{
			name:          "add rule to existing ingress with existing rule",
			existingRules: []networking.IngressRule{ruleHostFoo()},
			newRule:       ruleHostFoo(),
			expectedRules: []networking.IngressRule{ruleHostFoo()},
			err:           ErrorIngressRuleAlreadyExists,
		},
		{
			name:          "add rule to exiting ingress with tls secret",
			existingRules: []networking.IngressRule{ruleHostBar()},
			newRule:       ruleHostFoo(),
			expectedRules: []networking.IngressRule{ruleHostBar(), ruleHostFoo()},
			tlsSecret:     "my-secret",
			expectedTlsConfiguration: []networking.IngressTLS{{
				Hosts:      []string{ruleHostFoo().Host},
				SecretName: "my-secret",
			}},
		},
		{
			name:          "add rule to existing ingress with existing host with tls secret",
			existingRules: []networking.IngressRule{ruleHostFoo()},
			newRule:       ruleHostFoo2(),
			expectedRules: []networking.IngressRule{ruleHostFooTwoRules()},
			tlsSecret:     "my-secret",
			expectedTlsConfiguration: []networking.IngressTLS{{
				Hosts:      []string{ruleHostFoo().Host},
				SecretName: "my-secret",
			}},
		},
		{
			name:          "add rule to existing ingress with existing rule with tls secret",
			existingRules: []networking.IngressRule{ruleHostFoo()},
			newRule:       ruleHostFoo(),
			expectedRules: []networking.IngressRule{ruleHostFoo()},
			err:           ErrorIngressRuleAlreadyExists,
			tlsSecret:     "my-secret",
		},
		{
			name:          "add rule to existing ingress with existing rule with tls secret returns ErrorTlsConfigurationAlreadyExists",
			existingRules: []networking.IngressRule{ruleHostFoo()},
			newRule:       ruleHostFoo2(),
			expectedRules: []networking.IngressRule{ruleHostFoo()},
			err:           ErrorTlsConfigurationAlreadyExists,
			tlsSecret:     "my-secret",
			initialTlsConfiguration: []networking.IngressTLS{{
				Hosts:      []string{ruleHostFoo().Host},
				SecretName: "my-old-secret",
			}},
			expectedTlsConfiguration: []networking.IngressTLS{{
				Hosts:      []string{ruleHostFoo().Host},
				SecretName: "my-old-secret",
			}},
		},
		{
			name:          "add rule to existing ingress with existing rule with tls secret with same secret name",
			existingRules: []networking.IngressRule{ruleHostFoo()},
			newRule:       ruleHostFoo2(),
			expectedRules: []networking.IngressRule{ruleHostFooTwoRules()},
			err:           nil,
			tlsSecret:     "my-secret",
			initialTlsConfiguration: []networking.IngressTLS{{
				Hosts:      []string{ruleHostFoo().Host},
				SecretName: "my-secret",
			}},
			expectedTlsConfiguration: []networking.IngressTLS{{
				Hosts:      []string{ruleHostFoo().Host},
				SecretName: "my-secret",
			}},
		},
		{
			name:          "add rule to existing ingress with existing rule with existing tls secret",
			existingRules: []networking.IngressRule{ruleHostFoo()},
			newRule:       ruleHostBar(),
			expectedRules: []networking.IngressRule{ruleHostFoo(), ruleHostBar()},
			err:           nil,
			tlsSecret:     "my-secret",
			initialTlsConfiguration: []networking.IngressTLS{{
				Hosts:      []string{ruleHostFoo().Host},
				SecretName: "my-secret",
			}},
			expectedTlsConfiguration: []networking.IngressTLS{{
				Hosts:      []string{ruleHostFoo().Host, ruleHostBar().Host},
				SecretName: "my-secret",
			}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ingress := networking.Ingress{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec: networking.IngressSpec{
					Rules: test.existingRules,
					TLS:   test.initialTlsConfiguration,
				},
				Status: networking.IngressStatus{},
			}

			f := clienttesting.Fake{}
			f.AddReactor("get", "ingresses", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &ingress, nil
			})
			f.AddReactor("update", "ingresses", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
				spec := action.(clienttesting.UpdateAction).GetObject().(*networking.Ingress).Spec
				assert.Len(t, spec.Rules, len(test.expectedRules))
				assert.Equal(t, test.expectedRules, spec.Rules)
				assert.Equal(t, test.expectedTlsConfiguration, spec.TLS)
				return true, action.(clienttesting.UpdateAction).GetObject(), nil
			})

			ingressService := IngressService{
				kubeIngress: &fake.FakeIngresses{Fake: &fake.FakeNetworkingV1{Fake: &f}},
				ingressName: "foo",
			}

			created, err := ingressService.AddRule(context.TODO(), &test.newRule, test.tlsSecret)
			assert.Equal(t, test.err, err)
			assert.False(t, created)

			if err == nil {
				assert.Equal(t, test.expectedRules, ingress.Spec.Rules)
			}
			assert.Equal(t, test.expectedTlsConfiguration, ingress.Spec.TLS)
		})
	}
}

func TestIngressService_DeleteRule(t *testing.T) {
	tests := []struct {
		name              string
		inputRules        []networking.IngressRule
		serviceName       string
		servicePort       int32
		expectedRules     []networking.IngressRule
		expectedError     error
		initialTlsConfig  []networking.IngressTLS
		expectedTlsConfig []networking.IngressTLS
	}{
		{
			name:          "delete last rule by service name",
			inputRules:    []networking.IngressRule{ruleHostFoo()},
			serviceName:   "service-foo",
			servicePort:   0,
			expectedRules: []networking.IngressRule{},
			expectedError: nil,
		},
		{
			name:          "delete last rule by service name and port",
			inputRules:    []networking.IngressRule{ruleHostFoo()},
			serviceName:   "service-foo",
			servicePort:   80,
			expectedRules: []networking.IngressRule{},
			expectedError: nil,
		},
		{
			name:          "do not delete last rule by service name with invalid name",
			inputRules:    []networking.IngressRule{ruleHostFoo()},
			serviceName:   "service-bar",
			servicePort:   0,
			expectedRules: []networking.IngressRule{ruleHostFoo()},
			expectedError: ErrorIngressRuleNotFound,
		},
		{
			name:          "do not delete last rule by service name and port with invalid port",
			inputRules:    []networking.IngressRule{ruleHostFoo()},
			serviceName:   "service-foo",
			servicePort:   81,
			expectedRules: []networking.IngressRule{ruleHostFoo()},
			expectedError: ErrorIngressRuleNotFound,
		},
		{
			name:          "do not delete last rule by service name and port with invalid name",
			inputRules:    []networking.IngressRule{ruleHostFoo()},
			serviceName:   "service-bar",
			servicePort:   80,
			expectedRules: []networking.IngressRule{ruleHostFoo()},
			expectedError: ErrorIngressRuleNotFound,
		},
		// test cases for multiple existing rules
		{
			name:          "delete rule by service name",
			inputRules:    []networking.IngressRule{ruleHostFoo(), ruleHostBar()},
			serviceName:   "service-foo",
			servicePort:   0,
			expectedRules: []networking.IngressRule{ruleHostBar()},
			expectedError: nil,
		},
		{
			name:          "delete rule by service name and port",
			inputRules:    []networking.IngressRule{ruleHostFoo(), ruleHostBar()},
			serviceName:   "service-foo",
			servicePort:   80,
			expectedRules: []networking.IngressRule{ruleHostBar()},
			expectedError: nil,
		},
		{
			name:          "do not delete rule by service name",
			inputRules:    []networking.IngressRule{ruleHostFoo(), ruleHostBar()},
			serviceName:   "service-fooBar",
			servicePort:   0,
			expectedRules: []networking.IngressRule{ruleHostFoo(), ruleHostBar()},
			expectedError: ErrorIngressRuleNotFound,
		},
		{
			name:          "do not delete rule by service name and port with invalid name",
			inputRules:    []networking.IngressRule{ruleHostFoo(), ruleHostBar()},
			serviceName:   "service-fooBar",
			servicePort:   80,
			expectedRules: []networking.IngressRule{ruleHostFoo(), ruleHostBar()},
			expectedError: ErrorIngressRuleNotFound,
		},
		{
			name:          "do not delete rule by service name and port with invalid port",
			inputRules:    []networking.IngressRule{ruleHostFoo(), ruleHostBar()},
			serviceName:   "service-foo",
			servicePort:   81,
			expectedRules: []networking.IngressRule{ruleHostFoo(), ruleHostBar()},
			expectedError: ErrorIngressRuleNotFound,
		},
		{
			name:          "delete rule by service name multiple paths 1",
			inputRules:    []networking.IngressRule{ruleHostFooTwoRules(), ruleHostBar()},
			serviceName:   "service-foo",
			servicePort:   0,
			expectedRules: []networking.IngressRule{ruleHostFoo2(), ruleHostBar()},
			expectedError: nil,
		},
		{
			name:          "delete rule by service name multiple paths 2",
			inputRules:    []networking.IngressRule{ruleHostFooTwoRules(), ruleHostBar()},
			serviceName:   "service-foo-2",
			servicePort:   0,
			expectedRules: []networking.IngressRule{ruleHostFoo(), ruleHostBar()},
			expectedError: nil,
		},
		{
			name:          "delete rule by service name multiple paths 3",
			inputRules:    []networking.IngressRule{ruleHostFooTwoRules(), ruleHostBar()},
			serviceName:   "service-bar",
			servicePort:   0,
			expectedRules: []networking.IngressRule{ruleHostFooTwoRules()},
			expectedError: nil,
		},
		{
			name:          "delete rule by service name and port multiple paths",
			inputRules:    []networking.IngressRule{ruleHostFooTwoRules(), ruleHostBar()},
			serviceName:   "service-foo",
			servicePort:   80,
			expectedRules: []networking.IngressRule{ruleHostFoo2(), ruleHostBar()},
			expectedError: nil,
		},
		{
			name:          "delete rule by service name and port multiple paths 1",
			inputRules:    []networking.IngressRule{ruleHostFooTwoRules(), ruleHostBar()},
			serviceName:   "service-foo-2",
			servicePort:   80,
			expectedRules: []networking.IngressRule{ruleHostFoo(), ruleHostBar()},
			expectedError: nil,
		},
		{
			name:          "delete rule by service name and port multiple paths 3",
			inputRules:    []networking.IngressRule{ruleHostFooTwoRules(), ruleHostBar()},
			serviceName:   "service-bar",
			servicePort:   80,
			expectedRules: []networking.IngressRule{ruleHostFooTwoRules()},
			expectedError: nil,
		},
		{
			name:          "do not delete rule by service name multiple paths with invalid name",
			inputRules:    []networking.IngressRule{ruleHostFooTwoRules(), ruleHostBar()},
			serviceName:   "service-fooBar",
			servicePort:   0,
			expectedRules: []networking.IngressRule{ruleHostFooTwoRules(), ruleHostBar()},
			expectedError: ErrorIngressRuleNotFound,
		},
		{
			name:          "do not delete rule by service name multiple paths with invalid port",
			inputRules:    []networking.IngressRule{ruleHostFooTwoRules(), ruleHostBar()},
			serviceName:   "service-foo",
			servicePort:   81,
			expectedRules: []networking.IngressRule{ruleHostFooTwoRules(), ruleHostBar()},
			expectedError: ErrorIngressRuleNotFound,
		},
		//test cases for tls secrets
		{
			name:          "do delete tls configuration when host no longer exists (shared secret)",
			inputRules:    []networking.IngressRule{ruleHostFooTwoRules(), ruleHostBar()},
			serviceName:   "service-bar",
			servicePort:   0,
			expectedRules: []networking.IngressRule{ruleHostFooTwoRules()},
			expectedError: nil,
			initialTlsConfig: []networking.IngressTLS{{
				Hosts:      []string{"foo.com", "bar.com"},
				SecretName: "my-secret",
			}},
			expectedTlsConfig: []networking.IngressTLS{{
				Hosts:      []string{"foo.com"},
				SecretName: "my-secret",
			}},
		},
		{
			name:          "do delete tls configuration when host no longer exists (own secret)",
			inputRules:    []networking.IngressRule{ruleHostFooTwoRules(), ruleHostBar()},
			serviceName:   "service-bar",
			servicePort:   0,
			expectedRules: []networking.IngressRule{ruleHostFooTwoRules()},
			expectedError: nil,
			initialTlsConfig: []networking.IngressTLS{{
				Hosts:      []string{"foo.com"},
				SecretName: "my-secret",
			},
				{
					Hosts:      []string{"bar.com"},
					SecretName: "my-secret2",
				}},
			expectedTlsConfig: []networking.IngressTLS{{
				Hosts:      []string{"foo.com"},
				SecretName: "my-secret",
			}},
		},
		{
			name:          "do delete tls configuration when host no longer exists (last secret)",
			inputRules:    []networking.IngressRule{ruleHostFooTwoRules(), ruleHostBar()},
			serviceName:   "service-bar",
			servicePort:   0,
			expectedRules: []networking.IngressRule{ruleHostFooTwoRules()},
			expectedError: nil,
			initialTlsConfig: []networking.IngressTLS{{
				Hosts:      []string{"bar.com"},
				SecretName: "my-secret2",
			}},
			expectedTlsConfig: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ingress := networking.Ingress{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec: networking.IngressSpec{
					Rules: test.inputRules,
					TLS:   test.initialTlsConfig,
				},
				Status: networking.IngressStatus{},
			}

			isIngressDeleted := false

			f := clienttesting.Fake{}
			f.AddReactor("get", "ingresses", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &ingress, nil
			})
			f.AddReactor("delete", "ingresses", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
				isIngressDeleted = true
				return true, nil, nil
			})

			ingressService := IngressService{
				kubeIngress: &fake.FakeIngresses{Fake: &fake.FakeNetworkingV1{Fake: &f}},
				ingressName: "foo",
			}

			deleted, err := ingressService.DeleteRule(context.TODO(), test.serviceName, test.servicePort)
			assert.Equal(t, test.expectedError, err)

			if len(test.expectedRules) == 0 {
				assert.True(t, isIngressDeleted)
				assert.True(t, deleted)
			} else {
				assert.Equal(t, test.expectedRules, ingress.Spec.Rules)
				assert.Equal(t, test.expectedTlsConfig, ingress.Spec.TLS)
			}
		})
	}
}

func TestCreateIngressRule(t *testing.T) {
	assert.Equal(t, ruleHostFoo2(), *CreateIngressRule("foo.com", "/2", networking.PathTypePrefix, "service-foo-2", 80))
}

func ruleHostFoo() networking.IngressRule {
	pathType := networking.PathTypePrefix

	return networking.IngressRule{
		Host: "foo.com",
		IngressRuleValue: networking.IngressRuleValue{
			HTTP: &networking.HTTPIngressRuleValue{
				Paths: []networking.HTTPIngressPath{{
					Path:     "/",
					PathType: &pathType,
					Backend: networking.IngressBackend{
						Service: &networking.IngressServiceBackend{
							Name: "service-foo",
							Port: networking.ServiceBackendPort{
								Number: 80,
							},
						},
					},
				}},
			},
		},
	}
}

func ruleHostFoo2() networking.IngressRule {
	pathType := networking.PathTypePrefix

	return networking.IngressRule{
		Host: "foo.com",
		IngressRuleValue: networking.IngressRuleValue{
			HTTP: &networking.HTTPIngressRuleValue{
				Paths: []networking.HTTPIngressPath{{
					Path:     "/2",
					PathType: &pathType,
					Backend: networking.IngressBackend{
						Service: &networking.IngressServiceBackend{
							Name: "service-foo-2",
							Port: networking.ServiceBackendPort{
								Number: 80,
							},
						},
					},
				}},
			},
		},
	}
}

func ruleHostFooTwoRules() networking.IngressRule {
	pathType := networking.PathTypePrefix

	return networking.IngressRule{
		Host: "foo.com",
		IngressRuleValue: networking.IngressRuleValue{
			HTTP: &networking.HTTPIngressRuleValue{
				Paths: []networking.HTTPIngressPath{
					{
						Path:     "/",
						PathType: &pathType,
						Backend: networking.IngressBackend{
							Service: &networking.IngressServiceBackend{
								Name: "service-foo",
								Port: networking.ServiceBackendPort{
									Number: 80,
								},
							},
						},
					},
					{
						Path:     "/2",
						PathType: &pathType,
						Backend: networking.IngressBackend{
							Service: &networking.IngressServiceBackend{
								Name: "service-foo-2",
								Port: networking.ServiceBackendPort{
									Number: 80,
								},
							},
						},
					}},
			},
		},
	}
}

func ruleHostBar() networking.IngressRule {
	pathType := networking.PathTypePrefix

	return networking.IngressRule{
		Host: "bar.com",
		IngressRuleValue: networking.IngressRuleValue{
			HTTP: &networking.HTTPIngressRuleValue{
				Paths: []networking.HTTPIngressPath{{
					Path:     "/",
					PathType: &pathType,
					Backend: networking.IngressBackend{
						Service: &networking.IngressServiceBackend{
							Name: "service-bar",
							Port: networking.ServiceBackendPort{
								Number: 80,
							},
						},
					},
				}},
			},
		},
	}
}
