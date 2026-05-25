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
	"testing"

	"k8s.io/cloud-provider/fake"
)

func TestDeriveFromCloud_AllEnabled(t *testing.T) {
	cloud := &fake.Cloud{
		Provider: "test-provider",
	}

	caps := DeriveFromCloud(cloud)

	if caps.ProviderName() != "test-provider" {
		t.Errorf("expected provider name %q, got %q", "test-provider", caps.ProviderName())
	}

	expected := map[Capability]bool{
		CapLoadBalancer: true,
		CapInstances:    true,
		CapInstancesV2:  false,
		CapZones:        true,
		CapRoutes:       true,
		CapClusters:     true,
	}

	for cap, want := range expected {
		if got := caps.Has(cap); got != want {
			t.Errorf("capability %q: got %v, want %v", cap, got, want)
		}
	}
}

func TestDeriveFromCloud_DisabledFeatures(t *testing.T) {
	cloud := &fake.Cloud{
		Provider:             "limited-provider",
		DisableLoadBalancers: true,
		DisableRoutes:        true,
		DisableClusters:      true,
	}

	caps := DeriveFromCloud(cloud)

	if caps.Has(CapLoadBalancer) {
		t.Error("expected LoadBalancer to be disabled")
	}
	if caps.Has(CapRoutes) {
		t.Error("expected Routes to be disabled")
	}
	if caps.Has(CapClusters) {
		t.Error("expected Clusters to be disabled")
	}
	if !caps.Has(CapInstances) {
		t.Error("expected Instances to be enabled")
	}
	if !caps.Has(CapZones) {
		t.Error("expected Zones to be enabled")
	}
}

func TestDeriveFromCloud_InstancesV2Enabled(t *testing.T) {
	cloud := &fake.Cloud{
		Provider:          "v2-provider",
		EnableInstancesV2: true,
	}

	caps := DeriveFromCloud(cloud)

	if !caps.Has(CapInstancesV2) {
		t.Error("expected InstancesV2 to be enabled")
	}
	if !caps.Has(CapInstances) {
		t.Error("expected Instances to still be enabled")
	}
}

func TestDeriveFromCloud_AWSLike(t *testing.T) {
	cloud := &fake.Cloud{
		Provider:          "aws",
		EnableInstancesV2: true,
		DisableClusters:   true,
	}

	caps := DeriveFromCloud(cloud)
	caps.Caps[CapNodeDeletion] = true
	caps.Caps[CapSSHAccess] = true
	caps.Caps[CapInternalLoadBalancer] = true
	caps.Caps[CapVolumeProvisioning] = true

	if caps.Has(CapClusters) {
		t.Error("AWS should not support Clusters")
	}
	if !caps.Has(CapLoadBalancer) {
		t.Error("AWS should support LoadBalancer")
	}
	if !caps.Has(CapNodeDeletion) {
		t.Error("AWS should support NodeDeletion")
	}
	if !caps.Has(CapSSHAccess) {
		t.Error("AWS should support SSHAccess")
	}
	if !caps.Has(CapInternalLoadBalancer) {
		t.Error("AWS should support InternalLoadBalancer")
	}
}

func TestDeriveFromCloud_DefaultProviderName(t *testing.T) {
	cloud := &fake.Cloud{}
	caps := DeriveFromCloud(cloud)

	if caps.ProviderName() != "fake" {
		t.Errorf("expected default provider name %q, got %q", "fake", caps.ProviderName())
	}
}

func TestMapCapabilities_SubCapabilities(t *testing.T) {
	caps := &MapCapabilities{
		Name: "custom",
		Caps: map[Capability]bool{
			CapLoadBalancer: true,
			CapNodeDeletion: false,
		},
	}

	if !caps.Has(CapLoadBalancer) {
		t.Error("expected LoadBalancer to be true")
	}
	if caps.Has(CapNodeDeletion) {
		t.Error("expected NodeDeletion to be false")
	}
	if caps.Has(CapSSHAccess) {
		t.Error("expected unset capability to return false")
	}
}

func TestMapCapabilities_CustomCapability(t *testing.T) {
	customCap := Capability("mycloud/custom-feature")
	caps := &MapCapabilities{
		Name: "mycloud",
		Caps: map[Capability]bool{
			customCap: true,
		},
	}

	if !caps.Has(customCap) {
		t.Error("expected custom capability to be true")
	}
}

func TestRegistry(t *testing.T) {
	t.Cleanup(ResetCapabilities)

	if GetCapabilities("nonexistent") != nil {
		t.Error("expected nil for unregistered provider")
	}

	caps := &MapCapabilities{
		Name: "test",
		Caps: map[Capability]bool{CapLoadBalancer: true},
	}
	RegisterCapabilities("test", caps)

	got := GetCapabilities("test")
	if got == nil {
		t.Fatal("expected non-nil capabilities")
	}
	if !got.Has(CapLoadBalancer) {
		t.Error("expected LoadBalancer to be true")
	}
}

func TestRegistry_ReplaceExisting(t *testing.T) {
	t.Cleanup(ResetCapabilities)

	caps1 := &MapCapabilities{
		Name: "test",
		Caps: map[Capability]bool{CapLoadBalancer: true},
	}
	RegisterCapabilities("test", caps1)

	caps2 := &MapCapabilities{
		Name: "test",
		Caps: map[Capability]bool{CapLoadBalancer: false},
	}
	RegisterCapabilities("test", caps2)

	got := GetCapabilities("test")
	if got.Has(CapLoadBalancer) {
		t.Error("expected replacement to take effect")
	}
}
