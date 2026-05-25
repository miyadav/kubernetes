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
	"context"
	"errors"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	cloudprovider "k8s.io/cloud-provider"
)

// ConformanceConfig provides test fixtures for RunConformanceSuite.
// Providers populate fields relevant to their supported capabilities.
type ConformanceConfig struct {
	// Cloud is the cloudprovider.Interface implementation under test.
	Cloud cloudprovider.Interface

	// ClusterName is passed to methods that require it (LoadBalancer, Routes).
	ClusterName string

	// NodeName is a valid node name for Instances interface tests.
	// Leave empty if Instances is not supported.
	NodeName types.NodeName

	// ProviderID is a valid provider ID for Instances interface tests.
	// Leave empty if not applicable.
	ProviderID string

	// Node is a v1.Node object for InstancesV2 interface tests.
	// Leave nil if InstancesV2 is not supported.
	Node *v1.Node
}

// RunConformanceSuite runs provider-agnostic conformance tests against a
// cloudprovider.Interface implementation. It auto-derives capabilities from
// the cloud provider and only runs tests for supported capabilities.
//
// The suite validates that each declared sub-interface actually works —
// methods don't return cloudprovider.NotImplemented and return
// reasonable responses.
//
// Cloud providers call this from their own test suites:
//
//	func TestConformance(t *testing.T) {
//	    myCloud := newMyCloud(cfg)
//	    cloudprovidertesting.RunConformanceSuite(t, cloudprovidertesting.ConformanceConfig{
//	        Cloud:       myCloud,
//	        ClusterName: "test-cluster",
//	        NodeName:    "test-node-1",
//	        Node:        &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "test-node-1"}},
//	    })
//	}
func RunConformanceSuite(t *testing.T, cfg ConformanceConfig) {
	t.Helper()

	caps := DeriveFromCloud(cfg.Cloud)
	t.Logf("Running conformance suite for provider %q", caps.ProviderName())
	t.Logf("Derived capabilities: %+v", caps.Caps)

	t.Run("ProviderName", func(t *testing.T) {
		name := cfg.Cloud.ProviderName()
		if name == "" {
			t.Error("ProviderName() returned empty string")
		}
	})

	if caps.Has(CapLoadBalancer) {
		t.Run("LoadBalancer", func(t *testing.T) {
			runLoadBalancerConformance(t, cfg)
		})
	}

	if caps.Has(CapInstances) {
		t.Run("Instances", func(t *testing.T) {
			runInstancesConformance(t, cfg)
		})
	}

	if caps.Has(CapInstancesV2) {
		t.Run("InstancesV2", func(t *testing.T) {
			runInstancesV2Conformance(t, cfg)
		})
	}

	if caps.Has(CapZones) {
		t.Run("Zones", func(t *testing.T) {
			runZonesConformance(t, cfg)
		})
	}

	if caps.Has(CapRoutes) {
		t.Run("Routes", func(t *testing.T) {
			runRoutesConformance(t, cfg)
		})
	}

	if caps.Has(CapClusters) {
		t.Run("Clusters", func(t *testing.T) {
			runClustersConformance(t, cfg)
		})
	}
}

func runLoadBalancerConformance(t *testing.T, cfg ConformanceConfig) {
	lb, ok := cfg.Cloud.LoadBalancer()
	if !ok {
		t.Fatal("LoadBalancer() returned false but capability was declared")
	}

	ctx := context.Background()
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "conformance-test-svc",
			Namespace: "default",
			UID:       types.UID("conformance-test-uid"),
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeLoadBalancer,
			Ports: []v1.ServicePort{
				{Port: 80, Protocol: v1.ProtocolTCP},
			},
		},
	}

	t.Run("GetLoadBalancerName", func(t *testing.T) {
		name := lb.GetLoadBalancerName(ctx, cfg.ClusterName, svc)
		if name == "" {
			t.Error("GetLoadBalancerName returned empty string")
		}
	})

	t.Run("GetLoadBalancer", func(t *testing.T) {
		_, _, err := lb.GetLoadBalancer(ctx, cfg.ClusterName, svc)
		if err != nil && errors.Is(err, cloudprovider.NotImplemented) {
			t.Error("GetLoadBalancer returned NotImplemented for a declared LoadBalancer capability")
		}
	})
}

func runInstancesConformance(t *testing.T, cfg ConformanceConfig) {
	instances, ok := cfg.Cloud.Instances()
	if !ok {
		t.Fatal("Instances() returned false but capability was declared")
	}

	ctx := context.Background()

	t.Run("CurrentNodeName", func(t *testing.T) {
		nodeName, err := instances.CurrentNodeName(ctx, "test-hostname")
		if errors.Is(err, cloudprovider.NotImplemented) {
			t.Error("CurrentNodeName returned NotImplemented")
		}
		if err == nil && nodeName == "" {
			t.Error("CurrentNodeName returned empty node name without error")
		}
	})

	if cfg.NodeName != "" {
		t.Run("NodeAddresses", func(t *testing.T) {
			_, err := instances.NodeAddresses(ctx, cfg.NodeName)
			if errors.Is(err, cloudprovider.NotImplemented) {
				t.Error("NodeAddresses returned NotImplemented for a declared Instances capability")
			}
		})

		t.Run("InstanceID", func(t *testing.T) {
			_, err := instances.InstanceID(ctx, cfg.NodeName)
			if errors.Is(err, cloudprovider.NotImplemented) {
				t.Error("InstanceID returned NotImplemented for a declared Instances capability")
			}
		})

		t.Run("InstanceType", func(t *testing.T) {
			_, err := instances.InstanceType(ctx, cfg.NodeName)
			if errors.Is(err, cloudprovider.NotImplemented) {
				t.Error("InstanceType returned NotImplemented for a declared Instances capability")
			}
		})
	}

	if cfg.ProviderID != "" {
		t.Run("InstanceExistsByProviderID", func(t *testing.T) {
			_, err := instances.InstanceExistsByProviderID(ctx, cfg.ProviderID)
			if errors.Is(err, cloudprovider.NotImplemented) {
				t.Error("InstanceExistsByProviderID returned NotImplemented")
			}
		})
	}
}

func runInstancesV2Conformance(t *testing.T, cfg ConformanceConfig) {
	iv2, ok := cfg.Cloud.InstancesV2()
	if !ok {
		t.Fatal("InstancesV2() returned false but capability was declared")
	}

	if cfg.Node == nil {
		t.Skip("No Node fixture provided for InstancesV2 tests")
	}

	ctx := context.Background()

	t.Run("InstanceExists", func(t *testing.T) {
		_, err := iv2.InstanceExists(ctx, cfg.Node)
		if errors.Is(err, cloudprovider.NotImplemented) {
			t.Error("InstanceExists returned NotImplemented for a declared InstancesV2 capability")
		}
	})

	t.Run("InstanceShutdown", func(t *testing.T) {
		_, err := iv2.InstanceShutdown(ctx, cfg.Node)
		if errors.Is(err, cloudprovider.NotImplemented) {
			t.Error("InstanceShutdown returned NotImplemented for a declared InstancesV2 capability")
		}
	})

	t.Run("InstanceMetadata", func(t *testing.T) {
		metadata, err := iv2.InstanceMetadata(ctx, cfg.Node)
		if errors.Is(err, cloudprovider.NotImplemented) {
			t.Error("InstanceMetadata returned NotImplemented for a declared InstancesV2 capability")
		}
		if err == nil && metadata == nil {
			t.Error("InstanceMetadata returned nil metadata without error")
		}
	})
}

func runZonesConformance(t *testing.T, cfg ConformanceConfig) {
	zones, ok := cfg.Cloud.Zones()
	if !ok {
		t.Fatal("Zones() returned false but capability was declared")
	}

	ctx := context.Background()

	t.Run("GetZone", func(t *testing.T) {
		zone, err := zones.GetZone(ctx)
		if errors.Is(err, cloudprovider.NotImplemented) {
			t.Error("GetZone returned NotImplemented for a declared Zones capability")
		}
		if err == nil {
			if zone.FailureDomain == "" && zone.Region == "" {
				t.Error("GetZone returned Zone with empty FailureDomain and Region")
			}
		}
	})

	if cfg.NodeName != "" {
		t.Run("GetZoneByNodeName", func(t *testing.T) {
			_, err := zones.GetZoneByNodeName(ctx, cfg.NodeName)
			if errors.Is(err, cloudprovider.NotImplemented) {
				t.Error("GetZoneByNodeName returned NotImplemented")
			}
		})
	}

	if cfg.ProviderID != "" {
		t.Run("GetZoneByProviderID", func(t *testing.T) {
			_, err := zones.GetZoneByProviderID(ctx, cfg.ProviderID)
			if errors.Is(err, cloudprovider.NotImplemented) {
				t.Error("GetZoneByProviderID returned NotImplemented")
			}
		})
	}
}

func runRoutesConformance(t *testing.T, cfg ConformanceConfig) {
	routes, ok := cfg.Cloud.Routes()
	if !ok {
		t.Fatal("Routes() returned false but capability was declared")
	}

	ctx := context.Background()

	t.Run("ListRoutes", func(t *testing.T) {
		_, err := routes.ListRoutes(ctx, cfg.ClusterName)
		if errors.Is(err, cloudprovider.NotImplemented) {
			t.Error("ListRoutes returned NotImplemented for a declared Routes capability")
		}
	})
}

func runClustersConformance(t *testing.T, cfg ConformanceConfig) {
	clusters, ok := cfg.Cloud.Clusters()
	if !ok {
		t.Fatal("Clusters() returned false but capability was declared")
	}

	ctx := context.Background()

	t.Run("ListClusters", func(t *testing.T) {
		_, err := clusters.ListClusters(ctx)
		if errors.Is(err, cloudprovider.NotImplemented) {
			t.Error("ListClusters returned NotImplemented for a declared Clusters capability")
		}
	})
}
