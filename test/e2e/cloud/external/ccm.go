package external

import (
    "testing"
    "k8s.io/kubernetes/cloudprovider"
)

func TestInstanceLifecycle(t *testing.T) {
    // Inject cloud provider based on environment or config
    provider := GetCloudProvider()

    // Run common tests on instance lifecycle
    instance, err := provider.CreateInstance()
    if err != nil {
        t.Fatalf("failed to create instance: %v", err)
    }

    // Validate instance exists
    exists, err := provider.InstanceExists(instance)
    if err != nil || !exists {
        t.Fatalf("instance not found after creation")
    }

    // Cleanup
    if err := provider.DeleteInstance(instance); err != nil {
        t.Fatalf("failed to delete instance: %v", err)
    }
}

