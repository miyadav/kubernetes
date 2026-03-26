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
	"fmt"
	"sync"
)

var (
	// testDriverInstance holds the registered test driver
	testDriverInstance TestInterface
	// driverMutex protects access to testDriverInstance
	driverMutex sync.RWMutex
)

// RegisterTestDriver registers a TestInterface implementation for testing.
// Cloud providers should call this function to register their test driver,
// typically in an init() function in their cloud provider package.
//
// Example usage:
//
//	import "k8s.io/kubernetes/test/e2e/cloud/external"
//
//	func init() {
//	    external.RegisterTestDriver(&myCloudTestDriver{})
//	}
func RegisterTestDriver(driver TestInterface) error {
	driverMutex.Lock()
	defer driverMutex.Unlock()

	if testDriverInstance != nil {
		return fmt.Errorf("test driver already registered")
	}

	testDriverInstance = driver
	return nil
}

// GetTestDriver returns the registered TestInterface implementation.
// Returns nil if no driver has been registered.
func GetTestDriver() TestInterface {
	driverMutex.RLock()
	defer driverMutex.RUnlock()
	return testDriverInstance
}

// UnregisterTestDriver removes the currently registered test driver.
// This is primarily useful for testing.
func UnregisterTestDriver() {
	driverMutex.Lock()
	defer driverMutex.Unlock()
	testDriverInstance = nil
}
