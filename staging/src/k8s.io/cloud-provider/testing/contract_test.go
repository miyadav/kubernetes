/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testing

import (
	"fmt"
	"net"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/cloud-provider/fake"
)

func TestValidateCapabilities_Consistent(t *testing.T) {
	cloud := &fake.Cloud{Provider: "test"}
	caps := DeriveFromCloud(cloud)

	// DeriveFromCloud should produce consistent caps — no contract violations
	ValidateCapabilities(t, cloud, caps)
}

func TestValidateCapabilities_OverDeclared(t *testing.T) {
	cloud := &fake.Cloud{
		Provider:             "test",
		DisableLoadBalancers: true,
	}

	caps := &MapCapabilities{
		Name: "test",
		Caps: map[Capability]bool{
			CapLoadBalancer: true, // cloud says false, we say true → over-declared
		},
	}

	mockT := &mockTestingT{}
	ValidateCapabilities(mockT, cloud, caps)

	if !mockT.hadError {
		t.Error("expected contract violation for over-declared LoadBalancer")
	}
	if mockT.errorMsg == "" {
		t.Error("expected error message")
	}
}

func TestValidateCapabilities_UnderDeclared(t *testing.T) {
	cloud := &fake.Cloud{Provider: "test"}

	caps := &MapCapabilities{
		Name: "test",
		Caps: map[Capability]bool{
			CapLoadBalancer: false, // cloud says true, we say false → under-declared
		},
	}

	mockT := &mockTestingT{}
	ValidateCapabilities(mockT, cloud, caps)

	if !mockT.hadError {
		t.Error("expected contract violation for under-declared LoadBalancer")
	}
}

func TestValidateCapabilities_ProviderNameMismatch(t *testing.T) {
	cloud := &fake.Cloud{Provider: "aws"}
	caps := &MapCapabilities{
		Name: "gce", // mismatch
		Caps: map[Capability]bool{},
	}

	mockT := &mockTestingT{}
	ValidateCapabilities(mockT, cloud, caps)

	if !mockT.hadError {
		t.Error("expected error for provider name mismatch")
	}
}

func TestValidateCapabilities_AWSLike(t *testing.T) {
	cloud := &fake.Cloud{
		Provider:          "aws",
		EnableInstancesV2: true,
		DisableClusters:   true,
	}

	caps := DeriveFromCloud(cloud)
	// Should pass — derived caps are always consistent
	ValidateCapabilities(t, cloud, caps)
}

func TestConformanceSuite_FakeCloud(t *testing.T) {
	cloud := &fake.Cloud{
		Provider:          "fake",
		EnableInstancesV2: true,
		Exists:            true,
		ExistsByProviderID: true,
		ExternalIP:        net.ParseIP("1.2.3.4"),
		ExtID: map[types.NodeName]string{
			"test-node": "i-1234",
		},
		InstanceTypes: map[types.NodeName]string{
			"test-node": "m5.large",
		},
		ProviderID: map[types.NodeName]string{
			"test-node": "fake://i-1234",
		},
		Zone: cloudprovider.Zone{
			FailureDomain: "us-east-1a",
			Region:        "us-east-1",
		},
		RouteMap: map[string]*fake.Route{},
		ClusterList: []string{"test-cluster"},
	}

	RunConformanceSuite(t, ConformanceConfig{
		Cloud:       cloud,
		ClusterName: "test-cluster",
		NodeName:    "test-node",
		ProviderID:  "fake://i-1234",
		Node: &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
			},
			Spec: v1.NodeSpec{
				ProviderID: "fake://i-1234",
			},
		},
	})
}

func TestConformanceSuite_LimitedCloud(t *testing.T) {
	cloud := &fake.Cloud{
		Provider:             "limited",
		DisableLoadBalancers: true,
		DisableRoutes:        true,
		DisableClusters:      true,
		DisableInstances:     true,
		DisableZones:         true,
	}

	// Should pass — only runs ProviderName test since everything is disabled
	RunConformanceSuite(t, ConformanceConfig{
		Cloud:       cloud,
		ClusterName: "test-cluster",
	})
}

// mockTestingT captures errors from ValidateCapabilities without failing the real test.
type mockTestingT struct {
	hadError bool
	errorMsg string
}

func (m *mockTestingT) Helper()                                  {}
func (m *mockTestingT) Logf(format string, args ...interface{})  {}
func (m *mockTestingT) Run(name string, f func(*testing.T)) bool { return true }
func (m *mockTestingT) Errorf(format string, args ...interface{}) {
	m.hadError = true
	m.errorMsg = fmt.Sprintf(format, args...)
}
