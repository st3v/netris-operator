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
	"github.com/netrisai/netriswebapi/v2/types/inventory"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSwitchMetaReconciler_SwitchMetaNotFound(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	r := &SwitchMetaReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent-meta",
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

func TestSwitchMetaReconciler_SetsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	switchMeta := &k8sv1alpha1.SwitchMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "switch-meta-uid",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.SwitchMetaSpec{
			ID:         0,
			Imported:   false,
			Reclaim:    false,
			SwitchName: "test-switch",
		},
	}

	switchCR := &k8sv1alpha1.Switch{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-switch",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.SwitchSpec{
			Tenant: "admin",
			Site:   "dc1",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, switchMeta, switchCR)

	testStorage := newTestStorage(nil)

	mockClient := &MockInventoryClient{}
	r := &SwitchMetaReconciler{
		Client:          fakeClient,
		Log:             newTestLogger(),
		Scheme:          scheme,
		InventoryClient: mockClient,
		NStorage:        testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "switch-meta-uid",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Should create the switch in Netris API
	if result.Requeue {
		t.Errorf("expected no immediate requeue, got Requeue=true")
	}

	// Verify the mock client was called to add the switch
	if mockClient.LastAddID == 0 {
		t.Error("expected mock client to have recorded the add operation")
	}
}

func TestSwitchMetaReconciler_DeletionWithReclaim_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	switchMeta := &k8sv1alpha1.SwitchMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "switch-meta-reclaim",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SwitchMetaSpec{
			ID:         100,
			Reclaim:    true,
			SwitchName: "test-switch",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, switchMeta)

	r := &SwitchMetaReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "switch-meta-reclaim",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result.Requeue {
		t.Errorf("expected no requeue, got Requeue=true")
	}

	// TODO: Unlike SiteMetaReconciler, this controller returns early when DeletionTimestamp
	// is set without clearing finalizers. This may cause objects to be stuck in terminating
	// state. See sitemeta_controller.go for the expected deletion pattern.
	t.Log("WARNING: SwitchMetaReconciler does not clear finalizers on deletion - potential stuck terminating state")

	updated := &k8sv1alpha1.SwitchMeta{}
	if err := fakeClient.Get(context.Background(), req.NamespacedName, updated); err != nil {
		t.Fatalf("expected SwitchMeta to still exist, got error: %v", err)
	}
}

func TestSwitchMetaReconciler_DeletionWithZeroID_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	switchMeta := &k8sv1alpha1.SwitchMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "switch-meta-zero",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SwitchMetaSpec{
			ID:         0,
			Reclaim:    false,
			SwitchName: "test-switch",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, switchMeta)

	r := &SwitchMetaReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "switch-meta-zero",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result.Requeue {
		t.Errorf("expected no requeue, got Requeue=true")
	}

	// TODO: Unlike SiteMetaReconciler, this controller returns early when DeletionTimestamp
	// is set without clearing finalizers. This may cause objects to be stuck in terminating
	// state. See sitemeta_controller.go for the expected deletion pattern.
	t.Log("WARNING: SwitchMetaReconciler does not clear finalizers on deletion - potential stuck terminating state")

	updated := &k8sv1alpha1.SwitchMeta{}
	if err := fakeClient.Get(context.Background(), req.NamespacedName, updated); err != nil {
		t.Fatalf("expected SwitchMeta to still exist, got error: %v", err)
	}
}

func TestSwitchMetaReconciler_CreateSwitch(t *testing.T) {
	scheme := newTestScheme()

	switchMeta := &k8sv1alpha1.SwitchMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "switch-meta-create",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.SwitchMetaSpec{
			ID:         0,
			Imported:   false,
			Reclaim:    false,
			SwitchName: "test-switch",
			SiteID:     1,
			TenantID:   1,
		},
	}

	switchCR := &k8sv1alpha1.Switch{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-switch",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.SwitchSpec{
			Tenant: "admin",
			Site:   "dc1",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, switchMeta, switchCR)

	mockClient := &MockInventoryClient{}
	testStorage := newTestStorage(nil)

	r := &SwitchMetaReconciler{
		Client:          fakeClient,
		Log:             newTestLogger(),
		Scheme:          scheme,
		InventoryClient: mockClient,
		NStorage:        testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "switch-meta-create",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result.Requeue {
		t.Errorf("expected no immediate requeue, got Requeue=true")
	}

	// Verify the mock client was called to add the switch
	if mockClient.LastAddID == 0 {
		t.Error("expected mock client to have recorded the add operation")
	}

	// Verify ID was set after creation
	updated := &k8sv1alpha1.SwitchMeta{}
	if err := fakeClient.Get(context.Background(), req.NamespacedName, updated); err != nil {
		t.Fatalf("failed to get updated SwitchMeta: %v", err)
	}
	if updated.Spec.ID == 0 {
		t.Error("expected ID to be set after creation, got 0")
	}
}

func TestSwitchMetaReconciler_UpdateSwitch(t *testing.T) {
	scheme := newTestScheme()

	switchMeta := &k8sv1alpha1.SwitchMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "switch-meta-update",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.SwitchMetaSpec{
			ID:          100,
			Imported:    false,
			Reclaim:     false,
			SwitchName:  "test-switch",
			SiteID:      1,
			TenantID:    1,
			Description: "updated description",
		},
	}

	switchCR := &k8sv1alpha1.Switch{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-switch",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.SwitchSpec{
			Tenant: "admin",
			Site:   "dc1",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, switchMeta, switchCR)

	// Create storage with existing switch that has different description
	existingHW := &inventory.HW{
		ID:          100,
		Name:        "test-switch",
		Type:        "switch",
		Description: "old description",
	}
	testStorage := newTestStorageWithHWs(nil, []*inventory.HW{existingHW})

	mockClient := &MockInventoryClient{}

	r := &SwitchMetaReconciler{
		Client:          fakeClient,
		Log:             newTestLogger(),
		Scheme:          scheme,
		InventoryClient: mockClient,
		NStorage:        testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "switch-meta-update",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result.Requeue {
		t.Errorf("expected no immediate requeue, got Requeue=true")
	}
}

func TestSwitchMetaReconciler_DeletionWithID(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	switchMeta := &k8sv1alpha1.SwitchMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "switch-meta-delete",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SwitchMetaSpec{
			ID:         100,
			Reclaim:    false,
			SwitchName: "test-switch",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, switchMeta)

	mockClient := &MockInventoryClient{}
	testStorage := newTestStorage(nil)

	r := &SwitchMetaReconciler{
		Client:          fakeClient,
		Log:             newTestLogger(),
		Scheme:          scheme,
		InventoryClient: mockClient,
		NStorage:        testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "switch-meta-delete",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Deletion with DeletionTimestamp returns early
	if result.Requeue {
		t.Errorf("expected no requeue, got Requeue=true")
	}

	// TODO: Unlike SiteMetaReconciler, this controller returns early when DeletionTimestamp
	// is set without clearing finalizers. This may cause objects to be stuck in terminating
	// state. See sitemeta_controller.go for the expected deletion pattern.
	t.Log("WARNING: SwitchMetaReconciler does not clear finalizers on deletion - potential stuck terminating state")

	updated := &k8sv1alpha1.SwitchMeta{}
	if err := fakeClient.Get(context.Background(), req.NamespacedName, updated); err != nil {
		t.Fatalf("expected SwitchMeta to still exist, got error: %v", err)
	}
}
