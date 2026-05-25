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
	"sync"
)

var (
	registryMu   sync.RWMutex
	capabilities = make(map[string]TestCapabilities)
)

// RegisterCapabilities registers test capabilities for a named cloud provider.
// This is expected to be called from init() or provider setup, before tests run.
// Calling it more than once for the same provider replaces the previous registration.
func RegisterCapabilities(providerName string, caps TestCapabilities) {
	registryMu.Lock()
	defer registryMu.Unlock()
	capabilities[providerName] = caps
}

// GetCapabilities retrieves the registered test capabilities for a provider.
// Returns nil if no capabilities are registered for the provider.
func GetCapabilities(providerName string) TestCapabilities {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return capabilities[providerName]
}

// ResetCapabilities clears all registered capabilities. Intended for use in tests only.
func ResetCapabilities() {
	registryMu.Lock()
	defer registryMu.Unlock()
	capabilities = make(map[string]TestCapabilities)
}
