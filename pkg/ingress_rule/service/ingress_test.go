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

func TestIngressService_AddRule_ExistingIngress(t *testing.T) {
	ingress := networking.Ingress{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: networking.IngressSpec{
			Rules: []networking.IngressRule{ruleHostBar()},
		},
		Status: networking.IngressStatus{},
	}

	f := clienttesting.Fake{}
	f.AddReactor("get", "ingresses", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &ingress, nil
	})
	f.AddReactor("update", "ingresses", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		spec := action.(clienttesting.UpdateAction).GetObject().(*networking.Ingress).Spec
		assert.Len(t, spec.Rules, 2)

		assert.Equal(t, ruleHostBar(), spec.Rules[0])
		assert.Equal(t, ruleHostFoo(), spec.Rules[1])
		return true, action.(clienttesting.UpdateAction).GetObject(), nil
	})

	ingressService := IngressService{
		kubeIngress: &fake.FakeIngresses{Fake: &fake.FakeNetworkingV1{Fake: &f}},
		ingressName: "foo",
	}

	ruleFoo := ruleHostFoo()

	err := ingressService.AddRule(context.TODO(), &ruleFoo)
	assert.Nil(t, err)

	assert.Equal(t, ruleHostBar(), ingress.Spec.Rules[0])
	assert.Equal(t, ruleHostFoo(), ingress.Spec.Rules[1])
}

func TestIngressService_AddRule_ExistingIngress_ExistingHost(t *testing.T) {
	ingress := networking.Ingress{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: networking.IngressSpec{
			Rules: []networking.IngressRule{ruleHostFoo()},
		},
		Status: networking.IngressStatus{},
	}

	f := clienttesting.Fake{}
	f.AddReactor("get", "ingresses", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &ingress, nil
	})
	f.AddReactor("update", "ingresses", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		spec := action.(clienttesting.UpdateAction).GetObject().(*networking.Ingress).Spec
		assert.Len(t, spec.Rules, 1)

		assert.Equal(t, ruleHostFooTwoRules(), spec.Rules[0])
		return true, action.(clienttesting.UpdateAction).GetObject(), nil
	})

	ingressService := IngressService{
		kubeIngress: &fake.FakeIngresses{Fake: &fake.FakeNetworkingV1{Fake: &f}},
		ingressName: "foo",
	}

	ruleFoo := ruleHostFoo2()

	err := ingressService.AddRule(context.TODO(), &ruleFoo)
	assert.Nil(t, err)

	assert.Equal(t, ruleHostFooTwoRules(), ingress.Spec.Rules[0])
}

func TestIngressService_AddRule_ExistingIngress_ExistingRule(t *testing.T) {
	ingress := networking.Ingress{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: networking.IngressSpec{
			Rules: []networking.IngressRule{ruleHostBar()},
		},
		Status: networking.IngressStatus{},
	}

	f := clienttesting.Fake{}
	f.AddReactor("get", "ingresses", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &ingress, nil
	})

	ingressService := IngressService{
		kubeIngress: &fake.FakeIngresses{Fake: &fake.FakeNetworkingV1{Fake: &f}},
		ingressName: "foo",
	}

	rule := ruleHostBar()

	err := ingressService.AddRule(context.TODO(), &rule)
	assert.ErrorIs(t, err, ErrorIngressRuleAlreadyExists)

	assert.Equal(t, ruleHostBar(), ingress.Spec.Rules[0])
}

func TestIngressService_AddRule_NewIngress(t *testing.T) {
	var ingress *networking.Ingress

	f := clienttesting.Fake{}
	f.AddReactor("get", "ingresses", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, errors2.NewNotFound(action.GetResource().GroupResource(), action.(clienttesting.GetAction).GetName())
	})
	f.AddReactor("create", "ingresses", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		spec := action.(clienttesting.UpdateAction).GetObject().(*networking.Ingress).Spec
		assert.Len(t, spec.Rules, 1)

		assert.Equal(t, ruleHostFoo(), spec.Rules[0])

		ingress = action.(clienttesting.CreateAction).GetObject().(*networking.Ingress)
		return true, action.(clienttesting.CreateAction).GetObject(), nil
	})

	ingressService := IngressService{
		kubeIngress: &fake.FakeIngresses{Fake: &fake.FakeNetworkingV1{Fake: &f}},
		ingressName: "foo",
	}

	ruleFoo := ruleHostFoo()

	err := ingressService.AddRule(context.TODO(), &ruleFoo)
	assert.Nil(t, err)

	assert.Equal(t, ruleHostFoo(), ingress.Spec.Rules[0])
}

func TestIngressService_DeleteRule(t *testing.T) {
	tests := []struct {
		name          string
		inputRules    []networking.IngressRule
		serviceName   string
		servicePort   int32
		expectedRules []networking.IngressRule
		expectedError error
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
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ingress := networking.Ingress{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec: networking.IngressSpec{
					Rules: test.inputRules,
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

			err := ingressService.DeleteRule(context.TODO(), test.serviceName, test.servicePort)
			assert.Equal(t, test.expectedError, err)

			if len(test.expectedRules) == 0 {
				assert.True(t, isIngressDeleted)
			} else {
				assert.Equal(t, test.expectedRules, ingress.Spec.Rules)
			}
		})
	}
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
