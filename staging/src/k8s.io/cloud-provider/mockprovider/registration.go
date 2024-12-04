package mockprovider

import (
	"k8s.io/cloud-provider"
)

// RegisterMockProvider registers the mock provider.
func RegisterMockProvider() {
	cloudprovider.RegisterCloudProvider("mock", func() cloudprovider.CloudProvider {
		return NewMockProvider()
	})
}
