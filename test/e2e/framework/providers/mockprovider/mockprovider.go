/*
Copyright 2018 The Kubernetes Authors.

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

package mockprovider

import (
	"k8s.io/kubernetes/test/e2e/framework"
)

func init() {
	framework.RegisterProvider("mockprovider", newProvider)
}

func newProvider() (framework.ProviderInterface, error) {
	return &Provider{}, nil
}

// Provider is a structure to handle mock for e2e testing
// It does not do anything useful, it's there only to provide valid
// --provider=mockprovider cmdline option to allow testing .
type Provider struct {
	framework.NullProvider
}
