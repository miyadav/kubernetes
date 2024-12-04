package mockprovider

import (
	"testing"
)

func TestMockProvider_CreateCluster(t *testing.T) {
	provider := NewMockProvider()

	// Test the CreateCluster function
	err := provider.CreateCluster("test-cluster")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestMockProvider_DeleteCluster(t *testing.T) {
	provider := NewMockProvider()

	// Test the DeleteCluster function
	err := provider.DeleteCluster("test-cluster")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
