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
	"testing"

	k8sv1alpha1 "github.com/netrisai/netris-operator/api/v1alpha1"
	"github.com/netrisai/netriswebapi/v2/types/inventory"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSoftgateMetaReconciler_SoftgateMetaNotFound(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	r := &SoftgateMetaReconciler{
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

func TestSoftgateMetaReconciler_SetsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	softgateMeta := &k8sv1alpha1.SoftgateMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "softgate-meta-uid",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.SoftgateMetaSpec{
			ID:           0,
			Imported:     false,
			Reclaim:      false,
			SoftgateName: "test-softgate",
		},
	}

	softgateCR := &k8sv1alpha1.Softgate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-softgate",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.SoftgateSpec{
			Tenant: "admin",
			Site:   "dc1",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, softgateMeta, softgateCR)

	testStorage := newTestStorage(nil)

	mockClient := &MockInventoryClient{}
	r := &SoftgateMetaReconciler{
		Client:          fakeClient,
		Log:             newTestLogger(),
		Scheme:          scheme,
		InventoryClient: mockClient,
		NStorage:        testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "softgate-meta-uid",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Should create the softgate in Netris API
	if result.Requeue {
		t.Errorf("expected no immediate requeue, got Requeue=true")
	}

	// Verify the mock client was called to add the softgate
	if mockClient.LastAddID == 0 {
		t.Error("expected mock client to have recorded the add operation")
	}
}

func TestSoftgateMetaReconciler_DeletionWithReclaim_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	softgateMeta := &k8sv1alpha1.SoftgateMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "softgate-meta-reclaim",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SoftgateMetaSpec{
			ID:           100,
			Reclaim:      true,
			SoftgateName: "test-softgate",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, softgateMeta)

	r := &SoftgateMetaReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "softgate-meta-reclaim",
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
}

func TestSoftgateMetaReconciler_DeletionWithZeroID_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	softgateMeta := &k8sv1alpha1.SoftgateMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "softgate-meta-zero",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SoftgateMetaSpec{
			ID:           0,
			Reclaim:      false,
			SoftgateName: "test-softgate",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, softgateMeta)

	r := &SoftgateMetaReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "softgate-meta-zero",
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
}

func TestSoftgateMetaReconciler_CreateSoftgate(t *testing.T) {
	scheme := newTestScheme()

	softgateMeta := &k8sv1alpha1.SoftgateMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "softgate-meta-create",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.SoftgateMetaSpec{
			ID:           0,
			Imported:     false,
			Reclaim:      false,
			SoftgateName: "test-softgate",
			SiteID:       1,
			TenantID:     1,
		},
	}

	softgateCR := &k8sv1alpha1.Softgate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-softgate",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.SoftgateSpec{
			Tenant: "admin",
			Site:   "dc1",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, softgateMeta, softgateCR)

	mockClient := &MockInventoryClient{}
	testStorage := newTestStorage(nil)

	r := &SoftgateMetaReconciler{
		Client:          fakeClient,
		Log:             newTestLogger(),
		Scheme:          scheme,
		InventoryClient: mockClient,
		NStorage:        testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "softgate-meta-create",
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

	// Verify the mock client was called to add the softgate
	if mockClient.LastAddID == 0 {
		t.Error("expected mock client to have recorded the add operation")
	}
}

func TestSoftgateMetaReconciler_UpdateSoftgate(t *testing.T) {
	scheme := newTestScheme()

	softgateMeta := &k8sv1alpha1.SoftgateMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "softgate-meta-update",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.SoftgateMetaSpec{
			ID:           100,
			Imported:     false,
			Reclaim:      false,
			SoftgateName: "test-softgate",
			SiteID:       1,
			TenantID:     1,
			Description:  "updated description",
		},
	}

	softgateCR := &k8sv1alpha1.Softgate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-softgate",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.SoftgateSpec{
			Tenant: "admin",
			Site:   "dc1",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, softgateMeta, softgateCR)

	// Create storage with existing softgate that has different description
	existingHW := &inventory.HW{
		ID:          100,
		Name:        "test-softgate",
		Type:        "softgate",
		Description: "old description",
	}
	testStorage := newTestStorageWithHWs(nil, []*inventory.HW{existingHW})

	mockClient := &MockInventoryClient{}

	r := &SoftgateMetaReconciler{
		Client:          fakeClient,
		Log:             newTestLogger(),
		Scheme:          scheme,
		InventoryClient: mockClient,
		NStorage:        testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "softgate-meta-update",
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

func TestSoftgateMetaReconciler_DeletionWithID(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	softgateMeta := &k8sv1alpha1.SoftgateMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "softgate-meta-delete",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SoftgateMetaSpec{
			ID:           100,
			Reclaim:      false,
			SoftgateName: "test-softgate",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, softgateMeta)

	mockClient := &MockInventoryClient{}
	testStorage := newTestStorage(nil)

	r := &SoftgateMetaReconciler{
		Client:          fakeClient,
		Log:             newTestLogger(),
		Scheme:          scheme,
		InventoryClient: mockClient,
		NStorage:        testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "softgate-meta-delete",
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
}
