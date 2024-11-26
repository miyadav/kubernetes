package external

import (
    "fmt"
    "os"
    "k8s.io/kubernetes/cloudprovider"
    "k8s.io/kubernetes/cloudprovider/mock"
)

// GetCloudProvider returns either a mock provider or a real provider based on the config.
func GetCloudProvider() (cloudprovider.Interface, error) {
    if os.Getenv("USE_MOCK_PROVIDER") == "true" {
        return mock.NewMockProvider(), nil
    }

    return cloudprovider.GetCloudProvider()
}

