/*
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

package lbwatcher

import (
	"context"
	"fmt"
	"strings"
	"testing"

	k8sv1alpha1 "github.com/netrisai/netris-operator/api/v1alpha1"
	"github.com/netrisai/netris-operator/netrisstorage"
	"github.com/netrisai/netriswebapi/v2/types/ipam"
	"github.com/netrisai/netriswebapi/v2/types/site"
	"go.uber.org/zap/zapcore"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func init() {
	// Initialize logger and debugLogger for tests
	logger = zap.New(zap.UseDevMode(true))
	debugLogger = logger.V(int(zapcore.WarnLevel))
}

func newTestScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = k8sv1alpha1.AddToScheme(scheme)
	return scheme
}

func TestNewWatcher(t *testing.T) {
	tests := []struct {
		name    string
		storage *netrisstorage.Storage
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil storage returns error",
			storage: nil,
			wantErr: true,
			errMsg:  "Please provide NStorage",
		},
		{
			name:    "valid storage returns watcher",
			storage: &netrisstorage.Storage{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watcher, err := NewWatcher(tt.storage, nil, Options{})

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if err.Error() != tt.errMsg {
					t.Errorf("error = %q, want %q", err.Error(), tt.errMsg)
				}
				if watcher != nil {
					t.Errorf("expected nil watcher when error, got %v", watcher)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if watcher == nil {
					t.Errorf("expected watcher, got nil")
				}
				if watcher != nil && watcher.NStorage != tt.storage {
					t.Errorf("watcher.NStorage = %v, want %v", watcher.NStorage, tt.storage)
				}
			}
		})
	}
}

func TestFilterL4LBs(t *testing.T) {
	tests := []struct {
		name     string
		input    []k8sv1alpha1.L4LB
		expected int
	}{
		{
			name:     "empty list",
			input:    []k8sv1alpha1.L4LB{},
			expected: 0,
		},
		{
			name: "all LBs have service info",
			input: []k8sv1alpha1.L4LB{
				createL4LBWithServiceInfo("lb1", "svc1", "ns1", "uid1"),
				createL4LBWithServiceInfo("lb2", "svc2", "ns2", "uid2"),
			},
			expected: 2,
		},
		{
			name: "some LBs missing service info",
			input: []k8sv1alpha1.L4LB{
				createL4LBWithServiceInfo("lb1", "svc1", "ns1", "uid1"),
				createL4LBWithServiceInfo("lb2", "", "ns2", "uid2"),  // missing name
				createL4LBWithServiceInfo("lb3", "svc3", "", "uid3"), // missing namespace
				createL4LBWithServiceInfo("lb4", "svc4", "ns4", ""),  // missing uid
				createL4LBWithServiceInfo("lb5", "svc5", "ns5", "uid5"),
			},
			expected: 2,
		},
		{
			name: "all LBs missing service info",
			input: []k8sv1alpha1.L4LB{
				createL4LBWithServiceInfo("lb1", "", "", ""),
				createL4LBWithServiceInfo("lb2", "", "", ""),
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterL4LBs(tt.input)
			if len(result) != tt.expected {
				t.Errorf("filterL4LBs() returned %d items, want %d", len(result), tt.expected)
			}
		})
	}
}

func createL4LBWithServiceInfo(name, svcName, svcNamespace, svcUID string) k8sv1alpha1.L4LB {
	lb := k8sv1alpha1.L4LB{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   "default",
			Annotations: make(map[string]string),
		},
	}
	lb.SetServiceName(svcName)
	lb.SetServiceNamespace(svcNamespace)
	lb.SetServiceUID(svcUID)
	return lb
}

func TestCompareBackends(t *testing.T) {
	tests := []struct {
		name      string
		backends1 []k8sv1alpha1.L4LBBackend
		backends2 []k8sv1alpha1.L4LBBackend
		expected  bool
	}{
		{
			name:      "both empty",
			backends1: []k8sv1alpha1.L4LBBackend{},
			backends2: []k8sv1alpha1.L4LBBackend{},
			expected:  true,
		},
		{
			name:      "both nil",
			backends1: nil,
			backends2: nil,
			expected:  true,
		},
		{
			name:      "same backends same order",
			backends1: []k8sv1alpha1.L4LBBackend{"10.0.0.1:8080", "10.0.0.2:8080"},
			backends2: []k8sv1alpha1.L4LBBackend{"10.0.0.1:8080", "10.0.0.2:8080"},
			expected:  true,
		},
		{
			name:      "same backends different order",
			backends1: []k8sv1alpha1.L4LBBackend{"10.0.0.1:8080", "10.0.0.2:8080"},
			backends2: []k8sv1alpha1.L4LBBackend{"10.0.0.2:8080", "10.0.0.1:8080"},
			expected:  true,
		},
		{
			name:      "different backends",
			backends1: []k8sv1alpha1.L4LBBackend{"10.0.0.1:8080", "10.0.0.2:8080"},
			backends2: []k8sv1alpha1.L4LBBackend{"10.0.0.1:8080", "10.0.0.3:8080"},
			expected:  false,
		},
		{
			name:      "different lengths",
			backends1: []k8sv1alpha1.L4LBBackend{"10.0.0.1:8080"},
			backends2: []k8sv1alpha1.L4LBBackend{"10.0.0.1:8080", "10.0.0.2:8080"},
			expected:  false,
		},
		{
			name:      "one empty one has items",
			backends1: []k8sv1alpha1.L4LBBackend{},
			backends2: []k8sv1alpha1.L4LBBackend{"10.0.0.1:8080"},
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareBackends(tt.backends1, tt.backends2)
			if result != tt.expected {
				t.Errorf("compareBackends() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCompareLoadBalancers(t *testing.T) {
	tests := []struct {
		name            string
		existingLBs     []k8sv1alpha1.L4LB
		serviceLBs      []*k8sv1alpha1.L4LB
		wantCreateCount int
		wantUpdateCount int
		wantDeleteCount int
	}{
		{
			name:            "empty lists",
			existingLBs:     []k8sv1alpha1.L4LB{},
			serviceLBs:      []*k8sv1alpha1.L4LB{},
			wantCreateCount: 0,
			wantUpdateCount: 0,
			wantDeleteCount: 0,
		},
		{
			name:        "new service LB to create",
			existingLBs: []k8sv1alpha1.L4LB{},
			serviceLBs: []*k8sv1alpha1.L4LB{
				createL4LBPtr("new-lb", "svc1", "ns1", "uid1", "10.0.0.1"),
			},
			wantCreateCount: 1,
			wantUpdateCount: 0,
			wantDeleteCount: 0,
		},
		{
			name: "existing LB to delete",
			existingLBs: []k8sv1alpha1.L4LB{
				createL4LBWithServiceInfo("old-lb", "svc1", "ns1", "uid1"),
			},
			serviceLBs:      []*k8sv1alpha1.L4LB{},
			wantCreateCount: 0,
			wantUpdateCount: 0,
			wantDeleteCount: 1,
		},
		{
			name: "matching LB no update needed",
			existingLBs: []k8sv1alpha1.L4LB{
				createL4LBWithIP("matching-lb", "svc1", "ns1", "uid1", "10.0.0.1"),
			},
			serviceLBs: []*k8sv1alpha1.L4LB{
				createL4LBPtr("matching-lb", "svc1", "ns1", "uid1", "10.0.0.1"),
			},
			wantCreateCount: 0,
			wantUpdateCount: 0,
			wantDeleteCount: 0,
		},
		{
			name: "matching LB needs IP update",
			existingLBs: []k8sv1alpha1.L4LB{
				createL4LBWithIP("matching-lb", "svc1", "ns1", "uid1", "10.0.0.1"),
			},
			serviceLBs: []*k8sv1alpha1.L4LB{
				createL4LBPtr("matching-lb", "svc1", "ns1", "uid1", "10.0.0.2"),
			},
			wantCreateCount: 0,
			wantUpdateCount: 1,
			wantDeleteCount: 0,
		},
		{
			name: "matching LB needs timeout update",
			existingLBs: []k8sv1alpha1.L4LB{
				createL4LBWithTimeout("matching-lb", "svc1", "ns1", "uid1", "10.0.0.1", 30),
			},
			serviceLBs: []*k8sv1alpha1.L4LB{
				createL4LBPtrWithTimeout("matching-lb", "svc1", "ns1", "uid1", "10.0.0.1", 60),
			},
			wantCreateCount: 0,
			wantUpdateCount: 1,
			wantDeleteCount: 0,
		},
		{
			name: "matching LB needs backend update",
			existingLBs: []k8sv1alpha1.L4LB{
				createL4LBWithBackends("matching-lb", "svc1", "ns1", "uid1", "10.0.0.1",
					[]k8sv1alpha1.L4LBBackend{"10.0.0.100:8080"}),
			},
			serviceLBs: []*k8sv1alpha1.L4LB{
				createL4LBPtrWithBackends("matching-lb", "svc1", "ns1", "uid1", "10.0.0.1",
					[]k8sv1alpha1.L4LBBackend{"10.0.0.101:8080", "10.0.0.102:8080"}),
			},
			wantCreateCount: 0,
			wantUpdateCount: 1,
			wantDeleteCount: 0,
		},
		{
			name: "create and delete different LBs",
			existingLBs: []k8sv1alpha1.L4LB{
				createL4LBWithServiceInfo("old-lb", "svc1", "ns1", "uid1"),
			},
			serviceLBs: []*k8sv1alpha1.L4LB{
				createL4LBPtr("new-lb", "svc2", "ns2", "uid2", "10.0.0.2"),
			},
			wantCreateCount: 1,
			wantUpdateCount: 0,
			wantDeleteCount: 1,
		},
		{
			name: "new LB inherits IP from existing LB with same UID",
			existingLBs: []k8sv1alpha1.L4LB{
				createL4LBWithIP("lb-port-80", "svc1", "ns1", "uid1", "10.0.0.5"),
			},
			serviceLBs: []*k8sv1alpha1.L4LB{
				createL4LBPtr("lb-port-80", "svc1", "ns1", "uid1", "10.0.0.5"),
				createL4LBPtr("lb-port-443", "svc1", "ns1", "uid1", ""),
			},
			wantCreateCount: 1,
			wantUpdateCount: 0,
			wantDeleteCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toCreate, toUpdate, toDelete, _ := compareLoadBalancers(tt.existingLBs, tt.serviceLBs)

			if len(toCreate) != tt.wantCreateCount {
				t.Errorf("toCreate count = %d, want %d", len(toCreate), tt.wantCreateCount)
			}
			if len(toUpdate) != tt.wantUpdateCount {
				t.Errorf("toUpdate count = %d, want %d", len(toUpdate), tt.wantUpdateCount)
			}
			if len(toDelete) != tt.wantDeleteCount {
				t.Errorf("toDelete count = %d, want %d", len(toDelete), tt.wantDeleteCount)
			}
		})
	}
}

func createL4LBPtr(name, svcName, svcNamespace, svcUID, ip string) *k8sv1alpha1.L4LB {
	lb := &k8sv1alpha1.L4LB{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   "default",
			Annotations: make(map[string]string),
		},
		Spec: k8sv1alpha1.L4LBSpec{
			Frontend: k8sv1alpha1.L4LBFrontend{
				IP: ip,
			},
		},
	}
	lb.SetServiceName(svcName)
	lb.SetServiceNamespace(svcNamespace)
	lb.SetServiceUID(svcUID)
	return lb
}

func createL4LBWithIP(name, svcName, svcNamespace, svcUID, ip string) k8sv1alpha1.L4LB {
	lb := k8sv1alpha1.L4LB{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   "default",
			Annotations: make(map[string]string),
		},
		Spec: k8sv1alpha1.L4LBSpec{
			Frontend: k8sv1alpha1.L4LBFrontend{
				IP: ip,
			},
		},
	}
	lb.SetServiceName(svcName)
	lb.SetServiceNamespace(svcNamespace)
	lb.SetServiceUID(svcUID)
	return lb
}

func TestGetL4LBs(t *testing.T) {
	scheme := newTestScheme()

	tests := []struct {
		name        string
		existingLBs []k8sv1alpha1.L4LB
		wantCount   int
		wantErr     bool
	}{
		{
			name:        "no LBs exist",
			existingLBs: []k8sv1alpha1.L4LB{},
			wantCount:   0,
			wantErr:     false,
		},
		{
			name: "multiple LBs exist",
			existingLBs: []k8sv1alpha1.L4LB{
				{ObjectMeta: metav1.ObjectMeta{Name: "lb1", Namespace: "default"}},
				{ObjectMeta: metav1.ObjectMeta{Name: "lb2", Namespace: "default"}},
			},
			wantCount: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objects := make([]runtime.Object, len(tt.existingLBs))
			for i := range tt.existingLBs {
				objects[i] = &tt.existingLBs[i]
			}

			fakeClient := fake.NewFakeClientWithScheme(scheme, objects...)

			result, err := getL4LBs(fakeClient)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Fatalf("expected result, got nil")
				}
				if len(result.Items) != tt.wantCount {
					t.Errorf("got %d items, want %d", len(result.Items), tt.wantCount)
				}
			}
		})
	}
}

func TestDeleteL4LB(t *testing.T) {
	scheme := newTestScheme()

	lb := &k8sv1alpha1.L4LB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-lb",
			Namespace: "default",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, lb)

	// Verify it exists
	list := &k8sv1alpha1.L4LBList{}
	err := fakeClient.List(context.Background(), list)
	if err != nil {
		t.Fatalf("failed to list: %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(list.Items))
	}

	// Delete it
	err = deleteL4LB(fakeClient, *lb)
	if err != nil {
		t.Errorf("deleteL4LB() error = %v", err)
	}

	// Verify it's gone
	err = fakeClient.List(context.Background(), list)
	if err != nil {
		t.Fatalf("failed to list after delete: %v", err)
	}
	if len(list.Items) != 0 {
		t.Errorf("expected 0 items after delete, got %d", len(list.Items))
	}
}

func TestDeleteL4LB_NotFound(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	lb := k8sv1alpha1.L4LB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nonexistent-lb",
			Namespace: "default",
		},
	}

	// Should not error when deleting non-existent LB (uses IgnoreNotFound)
	err := deleteL4LB(fakeClient, lb)
	if err != nil {
		t.Errorf("deleteL4LB() should not error for non-existent LB, got %v", err)
	}
}

func TestCreateL4LB(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	lb := &k8sv1alpha1.L4LB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "new-lb",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.L4LBSpec{
			Frontend: k8sv1alpha1.L4LBFrontend{
				IP:   "10.0.0.1",
				Port: 80,
			},
		},
	}

	err := createL4LB(fakeClient, lb)
	if err != nil {
		t.Errorf("createL4LB() error = %v", err)
	}

	// Verify it was created
	list := &k8sv1alpha1.L4LBList{}
	err = fakeClient.List(context.Background(), list)
	if err != nil {
		t.Fatalf("failed to list: %v", err)
	}
	if len(list.Items) != 1 {
		t.Errorf("expected 1 item after create, got %d", len(list.Items))
	}
	if list.Items[0].Name != "new-lb" {
		t.Errorf("created LB name = %q, want %q", list.Items[0].Name, "new-lb")
	}
}

func TestUpdateL4LB(t *testing.T) {
	scheme := newTestScheme()

	lb := &k8sv1alpha1.L4LB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "existing-lb",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.L4LBSpec{
			Frontend: k8sv1alpha1.L4LBFrontend{
				IP:   "10.0.0.1",
				Port: 80,
			},
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, lb)

	// Update the LB
	lb.Spec.Frontend.IP = "10.0.0.2"
	err := updateL4LB(fakeClient, *lb)
	if err != nil {
		t.Errorf("updateL4LB() error = %v", err)
	}

	// Verify the update
	updated := &k8sv1alpha1.L4LB{}
	err = fakeClient.Get(context.Background(),
		types.NamespacedName{Name: "existing-lb", Namespace: "default"}, updated)
	if err != nil {
		t.Fatalf("failed to get updated LB: %v", err)
	}
	if updated.Spec.Frontend.IP != "10.0.0.2" {
		t.Errorf("updated IP = %q, want %q", updated.Spec.Frontend.IP, "10.0.0.2")
	}
}

func TestDeleteL4LBs(t *testing.T) {
	scheme := newTestScheme()

	lbs := []k8sv1alpha1.L4LB{
		{ObjectMeta: metav1.ObjectMeta{Name: "lb1", Namespace: "default"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "lb2", Namespace: "default"}},
	}

	objects := make([]runtime.Object, len(lbs))
	for i := range lbs {
		objects[i] = &lbs[i]
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, objects...)

	errors := deleteL4LBs(fakeClient, lbs)
	if len(errors) != 0 {
		t.Errorf("deleteL4LBs() returned %d errors, want 0", len(errors))
	}

	// Verify all deleted
	list := &k8sv1alpha1.L4LBList{}
	err := fakeClient.List(context.Background(), list)
	if err != nil {
		t.Fatalf("failed to list: %v", err)
	}
	if len(list.Items) != 0 {
		t.Errorf("expected 0 items after delete, got %d", len(list.Items))
	}
}

func TestCreateL4LBs(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	lbs := []*k8sv1alpha1.L4LB{
		{ObjectMeta: metav1.ObjectMeta{Name: "lb1", Namespace: "default"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "lb2", Namespace: "default"}},
	}

	ipAuto := map[string]string{}
	errors := createL4LBs(fakeClient, lbs, ipAuto)
	if len(errors) != 0 {
		t.Errorf("createL4LBs() returned %d errors, want 0", len(errors))
	}

	// Verify all created
	list := &k8sv1alpha1.L4LBList{}
	err := fakeClient.List(context.Background(), list)
	if err != nil {
		t.Fatalf("failed to list: %v", err)
	}
	if len(list.Items) != 2 {
		t.Errorf("expected 2 items after create, got %d", len(list.Items))
	}
}

func TestUpdateL4LBs(t *testing.T) {
	scheme := newTestScheme()

	lbs := []k8sv1alpha1.L4LB{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "lb1", Namespace: "default"},
			Spec:       k8sv1alpha1.L4LBSpec{Frontend: k8sv1alpha1.L4LBFrontend{IP: "10.0.0.1"}},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "lb2", Namespace: "default"},
			Spec:       k8sv1alpha1.L4LBSpec{Frontend: k8sv1alpha1.L4LBFrontend{IP: "10.0.0.2"}},
		},
	}

	objects := make([]runtime.Object, len(lbs))
	for i := range lbs {
		objects[i] = &lbs[i]
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, objects...)

	// Update IPs
	lbs[0].Spec.Frontend.IP = "10.0.0.10"
	lbs[1].Spec.Frontend.IP = "10.0.0.20"

	ipAuto := map[string]string{}
	errors := updateL4LBs(fakeClient, lbs, ipAuto)
	if len(errors) != 0 {
		t.Errorf("updateL4LBs() returned %d errors, want 0", len(errors))
	}
}

func createL4LBWithTimeout(name, svcName, svcNamespace, svcUID, ip string, timeout int) k8sv1alpha1.L4LB {
	lb := k8sv1alpha1.L4LB{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   "default",
			Annotations: make(map[string]string),
		},
		Spec: k8sv1alpha1.L4LBSpec{
			Frontend: k8sv1alpha1.L4LBFrontend{
				IP: ip,
			},
			Check: k8sv1alpha1.L4LBCheck{
				Timeout: timeout,
			},
		},
	}
	lb.SetServiceName(svcName)
	lb.SetServiceNamespace(svcNamespace)
	lb.SetServiceUID(svcUID)
	return lb
}

func createL4LBPtrWithTimeout(name, svcName, svcNamespace, svcUID, ip string, timeout int) *k8sv1alpha1.L4LB {
	lb := createL4LBWithTimeout(name, svcName, svcNamespace, svcUID, ip, timeout)
	return &lb
}

func createL4LBWithBackends(name, svcName, svcNamespace, svcUID, ip string, backends []k8sv1alpha1.L4LBBackend) k8sv1alpha1.L4LB {
	lb := k8sv1alpha1.L4LB{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   "default",
			Annotations: make(map[string]string),
		},
		Spec: k8sv1alpha1.L4LBSpec{
			Frontend: k8sv1alpha1.L4LBFrontend{
				IP: ip,
			},
			Backend: backends,
		},
	}
	lb.SetServiceName(svcName)
	lb.SetServiceNamespace(svcNamespace)
	lb.SetServiceUID(svcUID)
	return lb
}

func createL4LBPtrWithBackends(name, svcName, svcNamespace, svcUID, ip string, backends []k8sv1alpha1.L4LBBackend) *k8sv1alpha1.L4LB {
	lb := createL4LBWithBackends(name, svcName, svcNamespace, svcUID, ip, backends)
	return &lb
}

// mockK8sClient is a mock implementation of K8sClient for testing
type mockK8sClient struct {
	services         *v1.ServiceList
	service          *v1.Service
	updatedService   *v1.Service
	pods             *v1.PodList
	listServicesErr  error
	getServiceErr    error
	updateServiceErr error
	listPodsErr      error
}

func (m *mockK8sClient) ListServices(ctx context.Context, namespace string, opts metav1.ListOptions) (*v1.ServiceList, error) {
	return m.services, m.listServicesErr
}

func (m *mockK8sClient) GetService(ctx context.Context, namespace, name string, opts metav1.GetOptions) (*v1.Service, error) {
	return m.service, m.getServiceErr
}

func (m *mockK8sClient) UpdateServiceStatus(ctx context.Context, namespace string, service *v1.Service, opts metav1.UpdateOptions) (*v1.Service, error) {
	if m.updateServiceErr != nil {
		return nil, m.updateServiceErr
	}
	m.updatedService = service
	return service, nil
}

func (m *mockK8sClient) ListPods(ctx context.Context, namespace string, opts metav1.ListOptions) (*v1.PodList, error) {
	return m.pods, m.listPodsErr
}

func TestGetServices(t *testing.T) {
	tests := []struct {
		name      string
		mock      *mockK8sClient
		namespace string
		wantCount int
		wantErr   bool
	}{
		{
			name: "returns services successfully",
			mock: &mockK8sClient{
				services: &v1.ServiceList{
					Items: []v1.Service{
						{ObjectMeta: metav1.ObjectMeta{Name: "svc1", Namespace: "default"}},
						{ObjectMeta: metav1.ObjectMeta{Name: "svc2", Namespace: "default"}},
					},
				},
			},
			namespace: "",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "returns empty list",
			mock: &mockK8sClient{
				services: &v1.ServiceList{Items: []v1.Service{}},
			},
			namespace: "default",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "returns error",
			mock: &mockK8sClient{
				listServicesErr: fmt.Errorf("connection refused"),
			},
			namespace: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getServices(tt.mock, tt.namespace)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Fatal("expected result, got nil")
				}
				if len(result.Items) != tt.wantCount {
					t.Errorf("got %d services, want %d", len(result.Items), tt.wantCount)
				}
			}
		})
	}
}

func TestGetPodsByLabelSelector(t *testing.T) {
	tests := []struct {
		name      string
		mock      *mockK8sClient
		namespace string
		selectors string
		wantCount int
		wantErr   bool
	}{
		{
			name: "returns pods successfully",
			mock: &mockK8sClient{
				pods: &v1.PodList{
					Items: []v1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"},
							Status:     v1.PodStatus{HostIP: "10.0.0.1"},
						},
						{
							ObjectMeta: metav1.ObjectMeta{Name: "pod2", Namespace: "default"},
							Status:     v1.PodStatus{HostIP: "10.0.0.2"},
						},
					},
				},
			},
			namespace: "default",
			selectors: "app=nginx",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "returns empty list",
			mock: &mockK8sClient{
				pods: &v1.PodList{Items: []v1.Pod{}},
			},
			namespace: "default",
			selectors: "app=nonexistent",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "returns error",
			mock: &mockK8sClient{
				listPodsErr: fmt.Errorf("pods not found"),
			},
			namespace: "default",
			selectors: "app=nginx",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getPodsByLabelSelector(tt.mock, tt.namespace, tt.selectors)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Fatal("expected result, got nil")
				}
				if len(result.Items) != tt.wantCount {
					t.Errorf("got %d pods, want %d", len(result.Items), tt.wantCount)
				}
			}
		})
	}
}

func TestAssignIngress(t *testing.T) {
	tests := []struct {
		name      string
		mock      *mockK8sClient
		ips       []string
		namespace string
		svcName   string
		wantErr   bool
		wantIPs   []string
	}{
		{
			name: "assigns single IP successfully",
			mock: &mockK8sClient{
				service: &v1.Service{
					ObjectMeta: metav1.ObjectMeta{Name: "test-svc", Namespace: "default"},
				},
			},
			ips:       []string{"10.0.0.100"},
			namespace: "default",
			svcName:   "test-svc",
			wantErr:   false,
			wantIPs:   []string{"10.0.0.100"},
		},
		{
			name: "assigns multiple IPs successfully",
			mock: &mockK8sClient{
				service: &v1.Service{
					ObjectMeta: metav1.ObjectMeta{Name: "test-svc", Namespace: "default"},
				},
			},
			ips:       []string{"10.0.0.100", "10.0.0.101"},
			namespace: "default",
			svcName:   "test-svc",
			wantErr:   false,
			wantIPs:   []string{"10.0.0.100", "10.0.0.101"},
		},
		{
			name: "get service error",
			mock: &mockK8sClient{
				getServiceErr: fmt.Errorf("service not found"),
			},
			ips:       []string{"10.0.0.100"},
			namespace: "default",
			svcName:   "nonexistent",
			wantErr:   true,
		},
		{
			name: "update service error",
			mock: &mockK8sClient{
				service: &v1.Service{
					ObjectMeta: metav1.ObjectMeta{Name: "test-svc", Namespace: "default"},
				},
				updateServiceErr: fmt.Errorf("update failed"),
			},
			ips:       []string{"10.0.0.100"},
			namespace: "default",
			svcName:   "test-svc",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := assignIngress(tt.mock, tt.ips, tt.namespace, tt.svcName)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Fatal("expected result, got nil")
				}
				// Verify the ingress IPs were set
				if tt.mock.updatedService == nil {
					t.Fatal("expected updatedService to be set")
				}
				ingress := tt.mock.updatedService.Status.LoadBalancer.Ingress
				if len(ingress) != len(tt.wantIPs) {
					t.Errorf("got %d ingress IPs, want %d", len(ingress), len(tt.wantIPs))
				}
				for i, ip := range tt.wantIPs {
					if i < len(ingress) && ingress[i].IP != ip {
						t.Errorf("ingress[%d].IP = %q, want %q", i, ingress[i].IP, ip)
					}
				}
			}
		})
	}
}

func TestCreateEvent(t *testing.T) {
	tests := []struct {
		name      string
		mock      *mockK8sClient
		namespace string
		svcName   string
		reason    string
		message   string
		wantErr   bool
	}{
		{
			name: "creates event successfully",
			mock: &mockK8sClient{
				service: &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-svc",
						Namespace: "default",
						UID:       "test-uid",
					},
				},
			},
			namespace: "default",
			svcName:   "test-svc",
			reason:    "Failure",
			message:   "Load balancer failed",
			wantErr:   false,
		},
		{
			name: "service not found returns nil (ignored)",
			mock: &mockK8sClient{
				getServiceErr: fmt.Errorf("not found"),
			},
			namespace: "default",
			svcName:   "nonexistent",
			reason:    "Failure",
			message:   "Load balancer failed",
			wantErr:   true, // Error is wrapped and returned
		},
	}

	// Create a simple fake recorder for testing
	fakeRecorder := &fakeEventRecorder{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := createEvent(tt.mock, fakeRecorder, tt.namespace, tt.svcName, tt.reason, tt.message)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// fakeEventRecorder is a simple fake implementation of record.EventRecorder
type fakeEventRecorder struct {
	events []string
}

func (f *fakeEventRecorder) Event(object runtime.Object, eventtype, reason, message string) {
	f.events = append(f.events, fmt.Sprintf("%s/%s: %s", eventtype, reason, message))
}

func (f *fakeEventRecorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}) {
	f.events = append(f.events, fmt.Sprintf("%s/%s: %s", eventtype, reason, fmt.Sprintf(messageFmt, args...)))
}

func (f *fakeEventRecorder) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{}) {
	f.events = append(f.events, fmt.Sprintf("%s/%s: %s", eventtype, reason, fmt.Sprintf(messageFmt, args...)))
}

func TestGenerateLoadBalancers(t *testing.T) {
	tests := []struct {
		name        string
		mock        *mockK8sClient
		autoIPs     map[string]string
		lbTimeout   string
		storage     *netrisstorage.Storage
		wantCount   int
		wantErr     bool
		errContains string
	}{
		{
			name: "no services returns empty list",
			mock: &mockK8sClient{
				services: &v1.ServiceList{Items: []v1.Service{}},
			},
			autoIPs:   map[string]string{},
			lbTimeout: "2000",
			storage:   &netrisstorage.Storage{},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "non-LoadBalancer services are ignored",
			mock: &mockK8sClient{
				services: &v1.ServiceList{
					Items: []v1.Service{
						{
							ObjectMeta: metav1.ObjectMeta{Name: "cluster-svc", Namespace: "default"},
							Spec:       v1.ServiceSpec{Type: v1.ServiceTypeClusterIP},
						},
						{
							ObjectMeta: metav1.ObjectMeta{Name: "nodeport-svc", Namespace: "default"},
							Spec:       v1.ServiceSpec{Type: v1.ServiceTypeNodePort},
						},
					},
				},
			},
			autoIPs:   map[string]string{},
			lbTimeout: "2000",
			storage:   &netrisstorage.Storage{},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "LoadBalancer service with no pods returns empty",
			mock: &mockK8sClient{
				services: &v1.ServiceList{
					Items: []v1.Service{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "lb-svc",
								Namespace: "default",
								UID:       "test-uid-1",
							},
							Spec: v1.ServiceSpec{
								Type:     v1.ServiceTypeLoadBalancer,
								Selector: map[string]string{"app": "nginx"},
								Ports: []v1.ServicePort{
									{Name: "http", Port: 80, NodePort: 30080, Protocol: v1.ProtocolTCP},
								},
							},
						},
					},
				},
				pods: &v1.PodList{Items: []v1.Pod{}},
			},
			autoIPs:   map[string]string{},
			lbTimeout: "2000",
			storage:   &netrisstorage.Storage{},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "error getting services",
			mock: &mockK8sClient{
				listServicesErr: fmt.Errorf("connection refused"),
			},
			autoIPs:     map[string]string{},
			lbTimeout:   "2000",
			storage:     &netrisstorage.Storage{},
			wantErr:     true,
			errContains: "generateLoadBalancers",
		},
		{
			name: "invalid timeout",
			mock: &mockK8sClient{
				services: &v1.ServiceList{Items: []v1.Service{}},
			},
			autoIPs:     map[string]string{},
			lbTimeout:   "invalid",
			storage:     &netrisstorage.Storage{},
			wantErr:     true,
			errContains: "generateLoadBalancers",
		},
		{
			name: "error getting pods",
			mock: &mockK8sClient{
				services: &v1.ServiceList{
					Items: []v1.Service{
						{
							ObjectMeta: metav1.ObjectMeta{Name: "lb-svc", Namespace: "default"},
							Spec: v1.ServiceSpec{
								Type:     v1.ServiceTypeLoadBalancer,
								Selector: map[string]string{"app": "nginx"},
								Ports:    []v1.ServicePort{{Port: 80, NodePort: 30080}},
							},
						},
					},
				},
				listPodsErr: fmt.Errorf("pods error"),
			},
			autoIPs:     map[string]string{},
			lbTimeout:   "2000",
			storage:     &netrisstorage.Storage{},
			wantErr:     true,
			errContains: "generateLoadBalancers",
		},
		{
			name: "LoadBalancer service with pods but no ports",
			mock: &mockK8sClient{
				services: &v1.ServiceList{
					Items: []v1.Service{
						{
							ObjectMeta: metav1.ObjectMeta{Name: "lb-svc", Namespace: "default"},
							Spec: v1.ServiceSpec{
								Type:     v1.ServiceTypeLoadBalancer,
								Selector: map[string]string{"app": "nginx"},
								Ports:    []v1.ServicePort{},
							},
						},
					},
				},
				pods: &v1.PodList{
					Items: []v1.Pod{
						{
							ObjectMeta: metav1.ObjectMeta{Name: "pod1"},
							Status:     v1.PodStatus{HostIP: "10.0.0.1"},
						},
					},
				},
			},
			autoIPs:   map[string]string{},
			lbTimeout: "2000",
			storage:   &netrisstorage.Storage{},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Watcher{
				NStorage: tt.storage,
			}

			result, err := w.generateLoadBalancers(tt.mock, tt.autoIPs, tt.lbTimeout)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(result) != tt.wantCount {
					t.Errorf("got %d LBs, want %d", len(result), tt.wantCount)
				}
			}
		})
	}
}

func TestGenerateLoadBalancersWithValidService(t *testing.T) {
	// Create storage with subnets and sites for findSiteByIP
	// Note: findSiteByIP looks at subnet.Children, not top-level subnets
	subnetsStorage := netrisstorage.NewSubnetsStorage()
	subnetsStorage.Subnets = []*ipam.IPAM{
		{
			Prefix: "10.0.0.0/16",
			Children: []*ipam.IPAM{
				{
					Prefix: "10.0.0.0/24",
					Sites: []ipam.IDName{
						{ID: 1, Name: "site1"},
					},
				},
			},
		},
	}

	sitesStorage := netrisstorage.NewSitesStorage()
	sitesStorage.Sites = []*site.Site{
		{ID: 1, Name: "site1"},
	}

	storage := &netrisstorage.Storage{
		SubnetsStorage: subnetsStorage,
		SitesStorage:   sitesStorage,
	}

	mock := &mockK8sClient{
		services: &v1.ServiceList{
			Items: []v1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-lb-service",
						Namespace: "default",
						UID:       "uid-12345",
					},
					Spec: v1.ServiceSpec{
						Type:           v1.ServiceTypeLoadBalancer,
						LoadBalancerIP: "192.168.1.100",
						Selector:       map[string]string{"app": "nginx"},
						Ports: []v1.ServicePort{
							{Name: "http", Port: 80, NodePort: 30080, Protocol: v1.ProtocolTCP},
						},
					},
					Status: v1.ServiceStatus{
						LoadBalancer: v1.LoadBalancerStatus{
							Ingress: []v1.LoadBalancerIngress{
								{IP: "192.168.1.100"},
							},
						},
					},
				},
			},
		},
		pods: &v1.PodList{
			Items: []v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "nginx-pod-1"},
					Status:     v1.PodStatus{HostIP: "10.0.0.10"},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "nginx-pod-2"},
					Status:     v1.PodStatus{HostIP: "10.0.0.11"},
				},
			},
		},
	}

	w := &Watcher{NStorage: storage}

	result, err := w.generateLoadBalancers(mock, map[string]string{}, "2000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 LB, got %d", len(result))
	}

	lb := result[0]
	if lb.Spec.Frontend.IP != "192.168.1.100" {
		t.Errorf("frontend IP = %q, want %q", lb.Spec.Frontend.IP, "192.168.1.100")
	}
	if lb.Spec.Frontend.Port != 80 {
		t.Errorf("frontend port = %d, want %d", lb.Spec.Frontend.Port, 80)
	}
	if lb.Spec.Site != "site1" {
		t.Errorf("site = %q, want %q", lb.Spec.Site, "site1")
	}
	if len(lb.Spec.Backend) != 2 {
		t.Errorf("backend count = %d, want %d", len(lb.Spec.Backend), 2)
	}
	if lb.Spec.Check.Timeout != 2000 {
		t.Errorf("timeout = %d, want %d", lb.Spec.Check.Timeout, 2000)
	}
}

func TestGenerateLoadBalancersWithAutoIP(t *testing.T) {
	subnetsStorage := netrisstorage.NewSubnetsStorage()
	subnetsStorage.Subnets = []*ipam.IPAM{
		{
			Prefix: "10.0.0.0/16",
			Children: []*ipam.IPAM{
				{
					Prefix: "10.0.0.0/24",
					Sites:  []ipam.IDName{{ID: 1, Name: "site1"}},
				},
			},
		},
	}

	sitesStorage := netrisstorage.NewSitesStorage()
	sitesStorage.Sites = []*site.Site{{ID: 1, Name: "site1"}}

	storage := &netrisstorage.Storage{
		SubnetsStorage: subnetsStorage,
		SitesStorage:   sitesStorage,
	}

	mock := &mockK8sClient{
		services: &v1.ServiceList{
			Items: []v1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "auto-lb",
						Namespace: "default",
						UID:       "uid-auto",
					},
					Spec: v1.ServiceSpec{
						Type:     v1.ServiceTypeLoadBalancer,
						Selector: map[string]string{"app": "web"},
						Ports: []v1.ServicePort{
							{Port: 443, NodePort: 30443, Protocol: v1.ProtocolTCP},
						},
					},
				},
			},
		},
		pods: &v1.PodList{
			Items: []v1.Pod{
				{Status: v1.PodStatus{HostIP: "10.0.0.20"}},
			},
		},
	}

	w := &Watcher{NStorage: storage}
	autoIPs := map[string]string{"uid-auto": "172.16.0.50"}

	result, err := w.generateLoadBalancers(mock, autoIPs, "3000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 LB, got %d", len(result))
	}

	if result[0].Spec.Frontend.IP != "172.16.0.50" {
		t.Errorf("frontend IP = %q, want %q (from autoIPs)", result[0].Spec.Frontend.IP, "172.16.0.50")
	}
}

func TestDeleteL4LBsWithErrors(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	// Try to delete LBs that don't exist - should return no errors due to IgnoreNotFound
	lbs := []k8sv1alpha1.L4LB{
		{ObjectMeta: metav1.ObjectMeta{Name: "nonexistent1", Namespace: "default"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "nonexistent2", Namespace: "default"}},
	}

	errors := deleteL4LBs(fakeClient, lbs)
	if len(errors) != 0 {
		t.Errorf("deleteL4LBs() returned %d errors for nonexistent LBs, want 0", len(errors))
	}
}

func TestCreateL4LBsDuplicate(t *testing.T) {
	scheme := newTestScheme()

	existing := &k8sv1alpha1.L4LB{
		ObjectMeta: metav1.ObjectMeta{Name: "duplicate", Namespace: "default"},
	}
	fakeClient := fake.NewFakeClientWithScheme(scheme, existing)

	lbs := []*k8sv1alpha1.L4LB{
		{ObjectMeta: metav1.ObjectMeta{Name: "duplicate", Namespace: "default"}},
	}

	ipAuto := map[string]string{}
	errors := createL4LBs(fakeClient, lbs, ipAuto)
	if len(errors) != 1 {
		t.Errorf("createL4LBs() returned %d errors for duplicate, want 1", len(errors))
	}
}

func TestUpdateL4LBsNonexistent(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	lbs := []k8sv1alpha1.L4LB{
		{ObjectMeta: metav1.ObjectMeta{Name: "nonexistent", Namespace: "default"}},
	}

	ipAuto := map[string]string{}
	errors := updateL4LBs(fakeClient, lbs, ipAuto)
	if len(errors) != 1 {
		t.Errorf("updateL4LBs() returned %d errors for nonexistent, want 1", len(errors))
	}
}

func TestCreateEventServiceNotFound(t *testing.T) {
	mock := &mockK8sClient{
		getServiceErr: fmt.Errorf("services \"nonexistent\" not found"),
	}
	recorder := &fakeEventRecorder{}

	// The function should return error when service is not found
	err := createEvent(mock, recorder, "default", "nonexistent", "Failure", "test message")
	if err == nil {
		t.Error("expected error for not found service")
	}
}

func TestGetL4LBsError(t *testing.T) {
	scheme := newTestScheme()
	// Create a client that will fail - we can't easily make fake client fail,
	// but we can test the success path more thoroughly
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	result, err := getL4LBs(fakeClient)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}
