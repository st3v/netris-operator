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

func TestControllerMetaReconciler_ControllerMetaNotFound(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	r := &ControllerMetaReconciler{
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

func TestControllerMetaReconciler_SetsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	controllerMeta := &k8sv1alpha1.ControllerMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "controller-meta-uid",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.ControllerMetaSpec{
			ID:             0,
			Imported:       false,
			Reclaim:        false,
			ControllerName: "test-controller",
		},
	}

	controllerCR := &k8sv1alpha1.Controller{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-controller",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.ControllerSpec{
			Tenant: "admin",
			Site:   "dc1",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, controllerMeta, controllerCR)

	testStorage := newTestStorage(nil)

	mockClient := &MockInventoryClient{}
	r := &ControllerMetaReconciler{
		Client:          fakeClient,
		Log:             newTestLogger(),
		Scheme:          scheme,
		InventoryClient: mockClient,
		NStorage:        testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "controller-meta-uid",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Should create the controller in Netris API
	if result.Requeue {
		t.Errorf("expected no immediate requeue, got Requeue=true")
	}

	// Verify the mock client was called to add the controller
	if mockClient.LastAddID == 0 {
		t.Error("expected mock client to have recorded the add operation")
	}
}

func TestControllerMetaReconciler_DeletionWithReclaim_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	controllerMeta := &k8sv1alpha1.ControllerMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "controller-meta-reclaim",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.ControllerMetaSpec{
			ID:             100,
			Reclaim:        true,
			ControllerName: "test-controller",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, controllerMeta)

	r := &ControllerMetaReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "controller-meta-reclaim",
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

func TestControllerMetaReconciler_DeletionWithZeroID_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	controllerMeta := &k8sv1alpha1.ControllerMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "controller-meta-zero",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.ControllerMetaSpec{
			ID:             0,
			Reclaim:        false,
			ControllerName: "test-controller",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, controllerMeta)

	r := &ControllerMetaReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "controller-meta-zero",
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

func TestControllerMetaReconciler_CreateController(t *testing.T) {
	scheme := newTestScheme()

	controllerMeta := &k8sv1alpha1.ControllerMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "controller-meta-create",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.ControllerMetaSpec{
			ID:             0,
			Imported:       false,
			Reclaim:        false,
			ControllerName: "test-controller",
			SiteID:         1,
			TenantID:       1,
		},
	}

	controllerCR := &k8sv1alpha1.Controller{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-controller",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.ControllerSpec{
			Tenant: "admin",
			Site:   "dc1",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, controllerMeta, controllerCR)

	mockClient := &MockInventoryClient{}
	testStorage := newTestStorage(nil)

	r := &ControllerMetaReconciler{
		Client:          fakeClient,
		Log:             newTestLogger(),
		Scheme:          scheme,
		InventoryClient: mockClient,
		NStorage:        testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "controller-meta-create",
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

	// Verify the mock client was called to add the controller
	if mockClient.LastAddID == 0 {
		t.Error("expected mock client to have recorded the add operation")
	}
}

func TestControllerMetaReconciler_UpdateController(t *testing.T) {
	scheme := newTestScheme()

	controllerMeta := &k8sv1alpha1.ControllerMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "controller-meta-update",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.ControllerMetaSpec{
			ID:             100,
			Imported:       false,
			Reclaim:        false,
			ControllerName: "test-controller",
			SiteID:         1,
			TenantID:       1,
			Description:    "updated description",
		},
	}

	controllerCR := &k8sv1alpha1.Controller{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-controller",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.ControllerSpec{
			Tenant: "admin",
			Site:   "dc1",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, controllerMeta, controllerCR)

	// Create storage with existing controller that has different description
	existingHW := &inventory.HW{
		ID:          100,
		Name:        "test-controller",
		Type:        "controller",
		Description: "old description",
	}
	testStorage := newTestStorageWithHWs(nil, []*inventory.HW{existingHW})

	mockClient := &MockInventoryClient{}

	r := &ControllerMetaReconciler{
		Client:          fakeClient,
		Log:             newTestLogger(),
		Scheme:          scheme,
		InventoryClient: mockClient,
		NStorage:        testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "controller-meta-update",
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

func TestControllerMetaReconciler_DeletionWithID(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	controllerMeta := &k8sv1alpha1.ControllerMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "controller-meta-delete",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.ControllerMetaSpec{
			ID:             100,
			Reclaim:        false,
			ControllerName: "test-controller",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, controllerMeta)

	mockClient := &MockInventoryClient{}
	testStorage := newTestStorage(nil)

	r := &ControllerMetaReconciler{
		Client:          fakeClient,
		Log:             newTestLogger(),
		Scheme:          scheme,
		InventoryClient: mockClient,
		NStorage:        testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "controller-meta-delete",
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
