/*
Copyright 2015 The Kubernetes Authors.

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

package admission

import (
	"strconv"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestAdmitAndValidate(t *testing.T) {
	sysns := metav1.NamespaceSystem
	otherns := "default"
	tests := []struct {
		name      string
		ns        string
		operation Operation
		chain     chainAdmissionHandler
		accept    bool
		calls     map[string]bool
	}{
		{
			name:      "all accept",
			ns:        sysns,
			operation: Create,
			chain: []NamedHandler{
				makeNamedHandler("a", true, Update, Delete, Create),
				makeNamedHandler("b", true, Delete, Create),
				makeNamedHandler("c", true, Create),
			},
			calls:  map[string]bool{"a": true, "b": true, "c": true},
			accept: true,
		},
		{
			name:      "ignore handler",
			ns:        otherns,
			operation: Create,
			chain: []NamedHandler{
				makeNamedHandler("a", true, Update, Delete, Create),
				makeNamedHandler("b", false, Delete),
				makeNamedHandler("c", true, Create),
			},
			calls:  map[string]bool{"a": true, "c": true},
			accept: true,
		},
		{
			name:      "ignore all",
			ns:        sysns,
			operation: Connect,
			chain: []NamedHandler{
				makeNamedHandler("a", true, Update, Delete, Create),
				makeNamedHandler("b", false, Delete),
				makeNamedHandler("c", true, Create),
			},
			calls:  map[string]bool{},
			accept: true,
		},
		{
			name:      "reject one",
			ns:        otherns,
			operation: Delete,
			chain: []NamedHandler{
				makeNamedHandler("a", true, Update, Delete, Create),
				makeNamedHandler("b", false, Delete),
				makeNamedHandler("c", true, Create),
			},
			calls:  map[string]bool{"a": true, "b": true},
			accept: false,
		},
	}
	for _, test := range tests {
		Metrics.reset()
		t.Logf("testcase = %s", test.name)
		// call admit and check that validate was not called at all
		err := test.chain.Admit(NewAttributesRecord(nil, nil, schema.GroupVersionKind{}, test.ns, "", schema.GroupVersionResource{}, "", test.operation, nil))
		accepted := (err == nil)
		if accepted != test.accept {
			t.Errorf("unexpected result of admit call: %v", accepted)
		}
		for _, h := range test.chain {
			fake := h.Interface().(*FakeHandler)
			_, shouldBeCalled := test.calls[h.Name()]
			if shouldBeCalled != fake.admitCalled {
				t.Errorf("admit handler %s not called as expected: %v", h.Name(), fake.admitCalled)
				continue
			}
			if fake.validateCalled {
				t.Errorf("validate handler %s called during admit", h.Name())
			}

			//  reset value for validation test
			fake.admitCalled = false
		}

		labelFilter := map[string]string{
			"is_system_ns": strconv.FormatBool(test.ns == sysns),
			"type":         "mutating",
		}

		checkAdmitAndValidateMetrics(t, labelFilter, test.accept, test.calls)
		Metrics.reset()
		// call validate and check that admit was not called at all
		err = test.chain.Validate(NewAttributesRecord(nil, nil, schema.GroupVersionKind{}, test.ns, "", schema.GroupVersionResource{}, "", test.operation, nil))
		accepted = (err == nil)
		if accepted != test.accept {
			t.Errorf("unexpected result of validate call: %v\n", accepted)
		}
		for _, h := range test.chain {
			fake := h.Interface().(*FakeHandler)

			_, shouldBeCalled := test.calls[h.Name()]
			if shouldBeCalled != fake.validateCalled {
				t.Errorf("validate handler %s not called as expected: %v", h.Name(), fake.validateCalled)
				continue
			}

			if fake.admitCalled {
				t.Errorf("mutating handler unexpectedly called: %s", h.Name())
			}
		}

		labelFilter = map[string]string{
			"is_system_ns": strconv.FormatBool(test.ns == sysns),
			"type":         "validating",
		}

		checkAdmitAndValidateMetrics(t, labelFilter, test.accept, test.calls)
	}
}

func checkAdmitAndValidateMetrics(t *testing.T, labelFilter map[string]string, accept bool, calls map[string]bool) {
	acceptFilter := map[string]string{"rejected": "false"}
	for k, v := range labelFilter {
		acceptFilter[k] = v
	}

	rejectFilter := map[string]string{"rejected": "true"}
	for k, v := range labelFilter {
		rejectFilter[k] = v
	}

	if accept {
		// Ensure exactly one admission end-to-end admission accept should have been recorded.
		expectHistogramCountTotal(t, "apiserver_admission_step_latencies", acceptFilter, 1)

		// Ensure the expected count of admission controllers have been executed.
		expectHistogramCountTotal(t, "apiserver_admission_controller_latencies", acceptFilter, len(calls))
	} else {
		// When not accepted, ensure exactly one end-to-end rejection has been recorded.
		expectHistogramCountTotal(t, "apiserver_admission_step_latencies", rejectFilter, 1)
		if len(calls) > 0 {
			if len(calls) > 1 {
				// When not accepted, ensure that all but the last controller had been accepted, since
				// the chain stops execution at the first rejection.
				expectHistogramCountTotal(t, "apiserver_admission_controller_latencies", acceptFilter, len(calls)-1)
			}

			// When not accepted, ensure exactly one controller has been rejected.
			expectHistogramCountTotal(t, "apiserver_admission_controller_latencies", rejectFilter, 1)
		}
	}
}

func TestHandles(t *testing.T) {
	chain := chainAdmissionHandler{
		makeNamedHandler("a", true, Update, Delete, Create),
		makeNamedHandler("b", true, Delete, Create),
		makeNamedHandler("c", true, Create),
	}

	tests := []struct {
		name      string
		operation Operation
		chain     chainAdmissionHandler
		expected  bool
	}{
		{
			name:      "all handle",
			operation: Create,
			expected:  true,
		},
		{
			name:      "none handle",
			operation: Connect,
			expected:  false,
		},
		{
			name:      "some handle",
			operation: Delete,
			expected:  true,
		},
	}
	for _, test := range tests {
		handles := chain.Handles(test.operation)
		if handles != test.expected {
			t.Errorf("Unexpected handles result. Expected: %v. Actual: %v", test.expected, handles)
		}
	}
}
