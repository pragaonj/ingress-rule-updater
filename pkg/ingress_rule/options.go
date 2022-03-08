package ingress_rule

import (
	networking "k8s.io/api/networking/v1"
)

type Options struct {
	IngressName      string
	IngressClassName string
	Host             string
	Path             string
	Delete           bool
	Set              bool
	PathType         networking.PathType
	ServiceName      string
	PortNumber       int32
	TlsSecret        string
}
