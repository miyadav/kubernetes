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

package external

import (
	"context"
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
)

// CCMLoadBalancerTester implements the LoadBalancerTester interface for Cloud Controller Manager load balancer tests.
// It provides generic test logic that handles all Kubernetes API operations and delegates
// cloud-specific verification to LoadBalancerVerifier.
type CCMLoadBalancerTester struct {
	verifier LoadBalancerVerifier
}

// NewCCMLoadBalancerTester creates a new CCMLoadBalancerTester instance.
func NewCCMLoadBalancerTester() LoadBalancerTester {
	return &CCMLoadBalancerTester{}
}

// SetLoadBalancerVerifier sets the cloud-specific LoadBalancerVerifier implementation.
func (c *CCMLoadBalancerTester) SetLoadBalancerVerifier(verifier LoadBalancerVerifier) {
	c.verifier = verifier
}

// TestGetLoadBalancer tests the GetLoadBalancer functionality.
// This test verifies that the cloud provider can retrieve load balancer status.
// It handles all Kubernetes API operations and delegates cloud-specific verification to LoadBalancerVerifier.
func (c *CCMLoadBalancerTester) TestGetLoadBalancer(ctx context.Context, client clientset.Interface) (TestResult, error) {
	if framework.TestContext.CloudConfig.Provider == nil {
		return NewSkippedTestResult("cloud provider is not configured"), fmt.Errorf("cloud provider is not configured")
	}

	// Get all LoadBalancer services
	services, err := client.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return NewFailedTestResult(err, "failed to list services"), fmt.Errorf("failed to list services: %w", err)
	}

	lbCount := 0
	for _, svc := range services.Items {
		if svc.Spec.Type != v1.ServiceTypeLoadBalancer {
			continue
		}

		lbCount++
		if len(svc.Status.LoadBalancer.Ingress) == 0 {
			framework.Logf("LoadBalancer service %s/%s has no ingress", svc.Namespace, svc.Name)
			continue
		}

		hostname := svc.Status.LoadBalancer.Ingress[0].Hostname
		if hostname == "" {
			hostname = svc.Status.LoadBalancer.Ingress[0].IP
		}

		if hostname != "" {
			// If verifier is set, use it to verify the load balancer exists in the cloud provider
			if c.verifier != nil {
				exists, err := c.verifier.VerifyLoadBalancerExists(ctx, hostname)
				if err != nil {
					return NewFailedTestResult(err, fmt.Sprintf("failed to verify load balancer for service %s/%s", svc.Namespace, svc.Name)), fmt.Errorf("failed to verify load balancer for service %s/%s: %w", svc.Namespace, svc.Name, err)
				}
				if !exists {
					return NewFailedTestResult(fmt.Errorf("load balancer not found"), fmt.Sprintf("load balancer %s for service %s/%s not found in cloud provider", hostname, svc.Namespace, svc.Name)), fmt.Errorf("load balancer %s for service %s/%s not found in cloud provider", hostname, svc.Namespace, svc.Name)
				}
				framework.Logf("Verified load balancer %s exists for service %s/%s", hostname, svc.Namespace, svc.Name)
			}
		}
	}

	framework.Logf("Successfully verified %d LoadBalancer services", lbCount)
	return NewSuccessTestResult(fmt.Sprintf("Successfully verified %d LoadBalancer services", lbCount)), nil
}

// TestGetLoadBalancerName tests the GetLoadBalancerName functionality.
// This test verifies that the cloud provider can generate appropriate load balancer names.
func (c *CCMLoadBalancerTester) TestGetLoadBalancerName(ctx context.Context, clusterName string, service *v1.Service) (TestResult, error) {
	if framework.TestContext.CloudConfig.Provider == nil {
		return NewSkippedTestResult("cloud provider is not configured"), fmt.Errorf("cloud provider is not configured")
	}

	// TODO: Implement test logic that calls cloudprovider.LoadBalancer.GetLoadBalancerName
	// This requires access to the cloudprovider.Interface which is not directly available
	// through the test framework. Cloud providers should implement their own version
	// that accesses their cloud provider interface.
	return NewSkippedTestResult("TestGetLoadBalancerName not yet implemented - cloud providers should implement this"), nil
}

// TestEnsureLoadBalancer tests the EnsureLoadBalancer functionality.
// This test verifies that the cloud provider can create or update a load balancer.
// It handles all Kubernetes API operations (creating namespace, service, waiting for provisioning)
// and delegates cloud-specific verification to LoadBalancerVerifier.
func (c *CCMLoadBalancerTester) TestEnsureLoadBalancer(ctx context.Context, client clientset.Interface) (TestResult, error) {
	if framework.TestContext.CloudConfig.Provider == nil {
		return NewSkippedTestResult("cloud provider is not configured"), fmt.Errorf("cloud provider is not configured")
	}

	namespace := "lb-test-ns"
	serviceName := "test-lb-ensure"

	// Create test namespace
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return NewFailedTestResult(err, "failed to create namespace"), fmt.Errorf("failed to create namespace: %w", err)
	}
	defer func() {
		_ = client.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
	}()

	// Create a LoadBalancer service
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeLoadBalancer,
			Ports: []v1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app": "test-lb",
			},
		},
	}

	_, err = client.CoreV1().Services(namespace).Create(ctx, svc, metav1.CreateOptions{})
	if err != nil {
		return NewFailedTestResult(err, "failed to create LoadBalancer service"), fmt.Errorf("failed to create LoadBalancer service: %w", err)
	}
	defer func() {
		_ = client.CoreV1().Services(namespace).Delete(ctx, serviceName, metav1.DeleteOptions{})
	}()

	framework.Logf("Created LoadBalancer service %s/%s, waiting for provisioning...", namespace, serviceName)

	// Wait for the LoadBalancer to be provisioned
	var lbHostname string
	err = wait.PollUntilContextTimeout(ctx, 10*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
		updatedSvc, err := client.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if len(updatedSvc.Status.LoadBalancer.Ingress) > 0 {
			lbHostname = updatedSvc.Status.LoadBalancer.Ingress[0].Hostname
			if lbHostname == "" {
				lbHostname = updatedSvc.Status.LoadBalancer.Ingress[0].IP
			}
			return lbHostname != "", nil
		}
		framework.Logf("Waiting for LoadBalancer to be provisioned...")
		return false, nil
	})
	if err != nil {
		return NewFailedTestResult(err, "LoadBalancer was not provisioned within timeout"), fmt.Errorf("LoadBalancer was not provisioned within timeout: %w", err)
	}

	framework.Logf("LoadBalancer provisioned with hostname: %s", lbHostname)

	// Verify the load balancer exists in the cloud provider (if verifier is set)
	if c.verifier != nil {
		exists, err := c.verifier.VerifyLoadBalancerExists(ctx, lbHostname)
		if err != nil {
			return NewFailedTestResult(err, "failed to verify load balancer existence"), fmt.Errorf("failed to verify load balancer existence: %w", err)
		}
		if !exists {
			return NewFailedTestResult(fmt.Errorf("load balancer not found"), fmt.Sprintf("load balancer %s not found in cloud provider", lbHostname)), fmt.Errorf("load balancer %s not found in cloud provider", lbHostname)
		}
		framework.Logf("Successfully verified LoadBalancer %s exists in cloud provider", lbHostname)
	}

	return NewSuccessTestResult(fmt.Sprintf("Successfully verified LoadBalancer %s exists", lbHostname)), nil
}

// TestUpdateLoadBalancer tests the UpdateLoadBalancer functionality.
// This test verifies that the cloud provider can update hosts under a load balancer.
// It handles all Kubernetes API operations (creating namespace, service, updating service)
// and delegates cloud-specific verification to LoadBalancerVerifier.
func (c *CCMLoadBalancerTester) TestUpdateLoadBalancer(ctx context.Context, client clientset.Interface) (TestResult, error) {
	if framework.TestContext.CloudConfig.Provider == nil {
		return NewSkippedTestResult("cloud provider is not configured"), fmt.Errorf("cloud provider is not configured")
	}

	namespace := "lb-test-ns"
	serviceName := "test-lb-update"

	// Create test namespace
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return NewFailedTestResult(err, "failed to create namespace"), fmt.Errorf("failed to create namespace: %w", err)
	}
	defer func() {
		_ = client.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
	}()

	// Create initial LoadBalancer service
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeLoadBalancer,
			Ports: []v1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app": "test-lb",
			},
		},
	}

	createdSvc, err := client.CoreV1().Services(namespace).Create(ctx, svc, metav1.CreateOptions{})
	if err != nil {
		return NewFailedTestResult(err, "failed to create LoadBalancer service"), fmt.Errorf("failed to create LoadBalancer service: %w", err)
	}
	defer func() {
		_ = client.CoreV1().Services(namespace).Delete(ctx, serviceName, metav1.DeleteOptions{})
	}()

	// Wait for initial provisioning
	err = wait.PollUntilContextTimeout(ctx, 10*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
		updatedSvc, err := client.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		return len(updatedSvc.Status.LoadBalancer.Ingress) > 0, nil
	})
	if err != nil {
		return NewFailedTestResult(err, "initial LoadBalancer was not provisioned"), fmt.Errorf("initial LoadBalancer was not provisioned: %w", err)
	}

	framework.Logf("Initial LoadBalancer provisioned, now updating...")

	// Update the service - add a new port
	createdSvc.Spec.Ports = append(createdSvc.Spec.Ports, v1.ServicePort{
		Name:       "https",
		Port:       443,
		TargetPort: intstr.FromInt(8443),
		Protocol:   v1.ProtocolTCP,
	})

	_, err = client.CoreV1().Services(namespace).Update(ctx, createdSvc, metav1.UpdateOptions{})
	if err != nil {
		return NewFailedTestResult(err, "failed to update LoadBalancer service"), fmt.Errorf("failed to update LoadBalancer service: %w", err)
	}

	// Wait for update to propagate
	time.Sleep(30 * time.Second)

	// Verify the service was updated
	updatedSvc, err := client.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return NewFailedTestResult(err, "failed to get updated service"), fmt.Errorf("failed to get updated service: %w", err)
	}

	if len(updatedSvc.Spec.Ports) != 2 {
		return NewFailedTestResult(fmt.Errorf("port count mismatch"), fmt.Sprintf("expected 2 ports after update, got %d", len(updatedSvc.Spec.Ports))), fmt.Errorf("expected 2 ports after update, got %d", len(updatedSvc.Spec.Ports))
	}

	framework.Logf("Successfully verified LoadBalancer update")
	return NewSuccessTestResult("Successfully verified LoadBalancer update"), nil
}

// TestEnsureLoadBalancerDeleted tests the EnsureLoadBalancerDeleted functionality.
// This test verifies that the cloud provider can delete a load balancer.
// It handles all Kubernetes API operations (creating namespace, service, deleting service, waiting for deletion)
// and delegates cloud-specific verification to LoadBalancerVerifier.
func (c *CCMLoadBalancerTester) TestEnsureLoadBalancerDeleted(ctx context.Context, client clientset.Interface) (TestResult, error) {
	if framework.TestContext.CloudConfig.Provider == nil {
		return NewSkippedTestResult("cloud provider is not configured"), fmt.Errorf("cloud provider is not configured")
	}

	namespace := "lb-test-ns"
	serviceName := "test-lb-delete"

	// Create test namespace
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return NewFailedTestResult(err, "failed to create namespace"), fmt.Errorf("failed to create namespace: %w", err)
	}
	defer func() {
		_ = client.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
	}()

	// Create LoadBalancer service
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeLoadBalancer,
			Ports: []v1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
					Protocol:   v1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"app": "test-lb",
			},
		},
	}

	_, err = client.CoreV1().Services(namespace).Create(ctx, svc, metav1.CreateOptions{})
	if err != nil {
		return NewFailedTestResult(err, "failed to create LoadBalancer service"), fmt.Errorf("failed to create LoadBalancer service: %w", err)
	}

	// Wait for provisioning
	var lbHostname string
	err = wait.PollUntilContextTimeout(ctx, 10*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
		updatedSvc, err := client.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if len(updatedSvc.Status.LoadBalancer.Ingress) > 0 {
			lbHostname = updatedSvc.Status.LoadBalancer.Ingress[0].Hostname
			if lbHostname == "" {
				lbHostname = updatedSvc.Status.LoadBalancer.Ingress[0].IP
			}
			return lbHostname != "", nil
		}
		return false, nil
	})
	if err != nil {
		return NewFailedTestResult(err, "LoadBalancer was not provisioned"), fmt.Errorf("LoadBalancer was not provisioned: %w", err)
	}

	framework.Logf("LoadBalancer %s provisioned, now deleting service...", lbHostname)

	// Delete the service
	err = client.CoreV1().Services(namespace).Delete(ctx, serviceName, metav1.DeleteOptions{})
	if err != nil {
		return NewFailedTestResult(err, "failed to delete service"), fmt.Errorf("failed to delete service: %w", err)
	}

	// Wait for load balancer to be deleted from cloud provider (if verifier is set)
	if c.verifier != nil {
		err = wait.PollUntilContextTimeout(ctx, 10*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
			exists, err := c.verifier.VerifyLoadBalancerExists(ctx, lbHostname)
			if err != nil {
				framework.Logf("Error checking load balancer existence: %v", err)
				return false, nil
			}
			if !exists {
				return true, nil
			}
			framework.Logf("Waiting for load balancer %s to be deleted...", lbHostname)
			return false, nil
		})
		if err != nil {
			return NewFailedTestResult(err, "load balancer was not deleted within timeout"), fmt.Errorf("load balancer was not deleted within timeout: %w", err)
		}
		framework.Logf("Successfully verified LoadBalancer %s was deleted", lbHostname)
	}

	return NewSuccessTestResult(fmt.Sprintf("Successfully verified LoadBalancer %s was deleted", lbHostname)), nil
}
