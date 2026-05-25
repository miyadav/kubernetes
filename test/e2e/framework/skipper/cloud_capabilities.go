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

package skipper

import (
	cloudprovidertesting "k8s.io/cloud-provider/testing"
	"k8s.io/kubernetes/test/e2e/framework"
)

// SkipUnlessCloudCapability skips the current test if the cloud provider
// does not support the given capability.
//
// If no TestCapabilities have been registered for the current provider,
// the test is NOT skipped. This preserves backward compatibility: existing
// SkipUnlessProviderIs calls remain the authoritative gate until providers
// register capabilities. Once a provider registers, capability checks
// take effect and the provider-name checks can be removed.
func SkipUnlessCloudCapability(cap cloudprovidertesting.Capability) {
	providerName := framework.TestContext.Provider
	caps := cloudprovidertesting.GetCapabilities(providerName)
	if caps == nil {
		return
	}
	if !caps.Has(cap) {
		skipInternalf(1, "Cloud provider %q does not support capability %q", providerName, cap)
	}
}

// SkipIfCloudCapability skips the current test if the cloud provider
// DOES support the given capability. Useful for tests that exercise
// fallback behavior when a feature is absent.
func SkipIfCloudCapability(cap cloudprovidertesting.Capability) {
	providerName := framework.TestContext.Provider
	caps := cloudprovidertesting.GetCapabilities(providerName)
	if caps == nil {
		return
	}
	if caps.Has(cap) {
		skipInternalf(1, "Skipping because cloud provider %q supports capability %q", providerName, cap)
	}
}
