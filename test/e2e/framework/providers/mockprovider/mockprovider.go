package mockprovider

import (
	"k8s.io/kubernetes/test/e2e/framework"
)

func init() {
	framework.RegisterProvider("mockprovider", newProvider)
}

func newProvider() (framework.ProviderInterface, error) {
	return &Provider{}, nil
}

type Provider struct {
	framework.NullProvider
}
