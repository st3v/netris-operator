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

package controllers

import (
	"context"
	"testing"

	k8sv1alpha1 "github.com/netrisai/netris-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestAllocationReconciler_AllocationNotFound(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	r := &AllocationReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent-allocation",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.Requeue || result.RequeueAfter != 0 {
		t.Errorf("expected no requeue, got %v", result)
	}
}

func TestAllocationReconciler_SetsDefaultAnnotations(t *testing.T) {
	scheme := newTestScheme()

	allocation := &k8sv1alpha1.Allocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-allocation",
			Namespace: "default",
			UID:       "alloc-uid-123",
			Annotations: map[string]string{
				"test": "placeholder",
			},
		},
		Spec: k8sv1alpha1.AllocationSpec{
			Prefix: "10.0.0.0/24",
			Tenant: "admin",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, allocation)

	r := &AllocationReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-allocation",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.RequeueAfter != 0 {
		t.Errorf("expected no requeue delay after annotation update, got %v", result.RequeueAfter)
	}

	// Verify annotations were set
	updated := &k8sv1alpha1.Allocation{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated allocation: %v", err)
	}

	if updated.GetAnnotations()["resource.k8s.netris.ai/import"] != "false" {
		t.Errorf("expected import annotation 'false', got %q", updated.GetAnnotations()["resource.k8s.netris.ai/import"])
	}
	if updated.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"] != "delete" {
		t.Errorf("expected reclaimPolicy 'delete', got %q", updated.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"])
	}
}

func TestAllocationReconciler_SetsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	allocation := &k8sv1alpha1.Allocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-allocation",
			Namespace: "default",
			UID:       "alloc-uid-456",
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			// No finalizers
		},
		Spec: k8sv1alpha1.AllocationSpec{
			Prefix: "10.0.0.0/24",
			Tenant: "admin",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, allocation)

	r := &AllocationReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-allocation",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.RequeueAfter != 0 {
		t.Errorf("expected no requeue delay, got %v", result.RequeueAfter)
	}

	// Verify finalizer was set
	updated := &k8sv1alpha1.Allocation{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated allocation: %v", err)
	}

	if len(updated.GetFinalizers()) == 0 || updated.GetFinalizers()[0] != "resource.k8s.netris.ai/delete" {
		t.Errorf("expected finalizer to be set, got %v", updated.GetFinalizers())
	}
}

func TestAllocationReconciler_CreatesMeta(t *testing.T) {
	scheme := newTestScheme()

	allocation := &k8sv1alpha1.Allocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-allocation",
			Namespace:  "default",
			UID:        "alloc-uid-789",
			Generation: 1,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.AllocationSpec{
			Prefix: "192.168.0.0/16",
			Tenant: "engineering",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, allocation)

	r := &AllocationReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-allocation",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.RequeueAfter == 0 {
		t.Error("expected requeue with interval after creating meta")
	}

	// Verify AllocationMeta was created
	meta := &k8sv1alpha1.AllocationMeta{}
	metaKey := types.NamespacedName{
		Name:      string(allocation.UID),
		Namespace: "default",
	}
	err = fakeClient.Get(context.Background(), metaKey, meta)
	if err != nil {
		t.Fatalf("expected AllocationMeta to be created, got error: %v", err)
	}

	if meta.Spec.AllocationName != "test-allocation" {
		t.Errorf("expected AllocationName 'test-allocation', got %q", meta.Spec.AllocationName)
	}
	if meta.Spec.Prefix != "192.168.0.0/16" {
		t.Errorf("expected Prefix '192.168.0.0/16', got %q", meta.Spec.Prefix)
	}
	if meta.Spec.AllocationCRGeneration != 1 {
		t.Errorf("expected AllocationCRGeneration 1, got %d", meta.Spec.AllocationCRGeneration)
	}
}

func TestAllocationReconciler_MetaFoundNoChanges(t *testing.T) {
	scheme := newTestScheme()

	allocation := &k8sv1alpha1.Allocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-allocation",
			Namespace:  "default",
			UID:        "alloc-uid-abc",
			Generation: 1,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.AllocationSpec{
			Prefix: "10.0.0.0/8",
			Tenant: "admin",
		},
	}

	allocationMeta := &k8sv1alpha1.AllocationMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "alloc-uid-abc",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.AllocationMetaSpec{
			AllocationCRGeneration: 1, // matches
			Imported:               false,
			Reclaim:                false,
			AllocationName:         "test-allocation",
			Prefix:                 "10.0.0.0/8",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, allocation, allocationMeta)

	r := &AllocationReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-allocation",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.RequeueAfter == 0 {
		t.Error("expected requeue with interval")
	}
}

func TestAllocationReconciler_MetaFoundGenerationChanged(t *testing.T) {
	scheme := newTestScheme()

	allocation := &k8sv1alpha1.Allocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-allocation",
			Namespace:  "default",
			UID:        "alloc-uid-def",
			Generation: 2, // changed
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.AllocationSpec{
			Prefix: "172.16.0.0/12", // changed
			Tenant: "admin",
		},
	}

	allocationMeta := &k8sv1alpha1.AllocationMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "alloc-uid-def",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.AllocationMetaSpec{
			ID:                     42, // should be preserved
			AllocationCRGeneration: 1,  // old generation
			Imported:               false,
			Reclaim:                false,
			AllocationName:         "test-allocation",
			Prefix:                 "10.0.0.0/8", // old value
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, allocation, allocationMeta)

	r := &AllocationReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-allocation",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.RequeueAfter == 0 {
		t.Error("expected requeue with interval")
	}

	// Verify meta was updated
	updatedMeta := &k8sv1alpha1.AllocationMeta{}
	metaKey := types.NamespacedName{
		Name:      "alloc-uid-def",
		Namespace: "default",
	}
	err = fakeClient.Get(context.Background(), metaKey, updatedMeta)
	if err != nil {
		t.Fatalf("failed to get updated AllocationMeta: %v", err)
	}

	if updatedMeta.Spec.Prefix != "172.16.0.0/12" {
		t.Errorf("expected Prefix '172.16.0.0/12', got %q", updatedMeta.Spec.Prefix)
	}
	if updatedMeta.Spec.AllocationCRGeneration != 2 {
		t.Errorf("expected AllocationCRGeneration 2, got %d", updatedMeta.Spec.AllocationCRGeneration)
	}
	if updatedMeta.Spec.ID != 42 {
		t.Errorf("expected ID 42 to be preserved, got %d", updatedMeta.Spec.ID)
	}
}

func TestAllocationReconciler_DeletionNoMeta_ClearsFinalizer(t *testing.T) {
	// When deleted with no meta, should just clear finalizers (no API call)
	scheme := newTestScheme()

	now := metav1.Now()
	allocation := &k8sv1alpha1.Allocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-allocation",
			Namespace:         "default",
			UID:               "alloc-uid-nometa",
			DeletionTimestamp: &now,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.AllocationSpec{
			Prefix: "10.0.0.0/24",
			Tenant: "admin",
		},
	}

	// No AllocationMeta exists
	fakeClient := fake.NewFakeClientWithScheme(scheme, allocation)

	r := &AllocationReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
		// No Cred needed since no API call will be made
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-allocation",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Should complete deletion
	if result.RequeueAfter != 0 || result.Requeue {
		t.Errorf("expected no requeue after deletion, got %v", result)
	}

	// Verify finalizer was cleared
	updated := &k8sv1alpha1.Allocation{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated allocation: %v", err)
	}

	if len(updated.GetFinalizers()) != 0 {
		t.Errorf("expected finalizers to be cleared, got %v", updated.GetFinalizers())
	}
}

func TestAllocationReconciler_DeletionWithReclaim_SkipsAPICall(t *testing.T) {
	// When Reclaim=true, API delete should be skipped
	scheme := newTestScheme()

	now := metav1.Now()
	allocation := &k8sv1alpha1.Allocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-allocation",
			Namespace:         "default",
			UID:               "alloc-uid-reclaim",
			DeletionTimestamp: &now,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "true",
				"resource.k8s.netris.ai/reclaimPolicy": "retain",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.AllocationSpec{
			Prefix: "10.0.0.0/24",
			Tenant: "admin",
		},
	}

	allocationMeta := &k8sv1alpha1.AllocationMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "alloc-uid-reclaim",
			Namespace:  "default",
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.AllocationMetaSpec{
			ID:             100,
			Reclaim:        true, // This should skip the API call
			AllocationName: "test-allocation",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, allocation, allocationMeta)

	r := &AllocationReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
		// Cred is nil - if API call were attempted, it would panic
		// The fact that this test passes proves the API call was skipped
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-allocation",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Deletion should complete without error (API call was skipped due to Reclaim=true)
	if result.Requeue {
		t.Errorf("expected no requeue, got Requeue=true")
	}
}

func TestAllocationReconciler_ImportedAllocation(t *testing.T) {
	scheme := newTestScheme()

	allocation := &k8sv1alpha1.Allocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "imported-allocation",
			Namespace:  "default",
			UID:        "alloc-uid-import",
			Generation: 1,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "true",
				"resource.k8s.netris.ai/reclaimPolicy": "retain",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.AllocationSpec{
			Prefix: "10.0.0.0/8",
			Tenant: "admin",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, allocation)

	r := &AllocationReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "imported-allocation",
			Namespace: "default",
		},
	}

	_, err := r.Reconcile(req)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify AllocationMeta was created with correct flags
	meta := &k8sv1alpha1.AllocationMeta{}
	metaKey := types.NamespacedName{
		Name:      string(allocation.UID),
		Namespace: "default",
	}
	err = fakeClient.Get(context.Background(), metaKey, meta)
	if err != nil {
		t.Fatalf("expected AllocationMeta to be created, got error: %v", err)
	}

	if !meta.Spec.Imported {
		t.Error("expected Imported to be true")
	}
	if !meta.Spec.Reclaim {
		t.Error("expected Reclaim to be true")
	}
}
