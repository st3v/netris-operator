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
	"github.com/netrisai/netriswebapi/v1/types/tenant"
	"github.com/netrisai/netriswebapi/v2/types/site"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrlRuntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestControllerReconciler_ControllerNotFound(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	r := &ControllerReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrlRuntime.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent-controller",
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

func TestControllerReconciler_SetsDefaultAnnotations(t *testing.T) {
	scheme := newTestScheme()

	controller := &k8sv1alpha1.Controller{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-controller",
			Namespace: "default",
			UID:       "controller-uid-123",
			Annotations: map[string]string{
				"test": "placeholder",
			},
		},
		Spec: k8sv1alpha1.ControllerSpec{
			Tenant: "admin",
			Site:   "dc1",
			MainIP: "10.0.0.1",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, controller)

	r := &ControllerReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrlRuntime.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-controller",
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

	updated := &k8sv1alpha1.Controller{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated Controller: %v", err)
	}

	if updated.GetAnnotations()["resource.k8s.netris.ai/import"] != "false" {
		t.Errorf("expected import annotation 'false', got %q", updated.GetAnnotations()["resource.k8s.netris.ai/import"])
	}
	if updated.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"] != "delete" {
		t.Errorf("expected reclaimPolicy 'delete', got %q", updated.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"])
	}
}

func TestControllerReconciler_SetsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	controller := &k8sv1alpha1.Controller{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-controller",
			Namespace: "default",
			UID:       "controller-uid-456",
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
		},
		Spec: k8sv1alpha1.ControllerSpec{
			Tenant: "admin",
			Site:   "dc1",
			MainIP: "10.0.0.1",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, controller)

	r := &ControllerReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrlRuntime.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-controller",
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

	updated := &k8sv1alpha1.Controller{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated Controller: %v", err)
	}

	if len(updated.GetFinalizers()) == 0 || updated.GetFinalizers()[0] != "resource.k8s.netris.ai/delete" {
		t.Errorf("expected finalizer to be set, got %v", updated.GetFinalizers())
	}
}

func TestControllerReconciler_MetaFoundNoChanges(t *testing.T) {
	scheme := newTestScheme()

	controller := &k8sv1alpha1.Controller{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-controller",
			Namespace:  "default",
			UID:        "controller-uid-abc",
			Generation: 1,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.ControllerSpec{
			Tenant: "admin",
			Site:   "dc1",
			MainIP: "10.0.0.1",
		},
	}

	controllerMeta := &k8sv1alpha1.ControllerMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "controller-uid-abc",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.ControllerMetaSpec{
			ControllerCRGeneration: 1,
			Imported:               false,
			Reclaim:                false,
			ControllerName:         "test-controller",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, controller, controllerMeta)

	r := &ControllerReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrlRuntime.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-controller",
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

func TestControllerReconciler_DeletionNoMeta_ClearsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	controller := &k8sv1alpha1.Controller{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-controller",
			Namespace:         "default",
			UID:               "controller-uid-nometa",
			DeletionTimestamp: &now,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.ControllerSpec{
			Tenant: "admin",
			Site:   "dc1",
			MainIP: "10.0.0.1",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, controller)

	r := &ControllerReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrlRuntime.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-controller",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result.RequeueAfter != 0 || result.Requeue {
		t.Errorf("expected no requeue after deletion, got %v", result)
	}

	updated := &k8sv1alpha1.Controller{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated Controller: %v", err)
	}

	if len(updated.GetFinalizers()) != 0 {
		t.Errorf("expected finalizers to be cleared, got %v", updated.GetFinalizers())
	}
}

func TestControllerReconciler_DeletionWithReclaim_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	controller := &k8sv1alpha1.Controller{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-controller",
			Namespace:         "default",
			UID:               "controller-uid-reclaim",
			DeletionTimestamp: &now,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "true",
				"resource.k8s.netris.ai/reclaimPolicy": "retain",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.ControllerSpec{
			Tenant: "admin",
			Site:   "dc1",
			MainIP: "10.0.0.1",
		},
	}

	controllerMeta := &k8sv1alpha1.ControllerMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "controller-uid-reclaim",
			Namespace:  "default",
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.ControllerMetaSpec{
			ID:             100,
			Reclaim:        true,
			ControllerName: "test-controller",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, controller, controllerMeta)

	r := &ControllerReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrlRuntime.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-controller",
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

func TestControllerReconciler_CreatesMeta(t *testing.T) {
	scheme := newTestScheme()

	ctrl := &k8sv1alpha1.Controller{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-controller",
			Namespace:  "default",
			UID:        "controller-uid-createsmeta",
			Generation: 1,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.ControllerSpec{
			Tenant: "admin",
			Site:   "dc1",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, ctrl)

	// Create test storage with sites and tenants
	testStorage := newTestStorage([]*site.Site{
		{ID: 1, Name: "dc1"},
	})
	testStorage.TenantsStorage.Tenants = []*tenant.Tenant{
		{ID: 1, Name: "admin"},
	}

	r := &ControllerReconciler{
		Client:          fakeClient,
		Log:             newTestLogger(),
		Scheme:          scheme,
		InventoryClient: &MockInventoryClient{},
		NStorage:        testStorage,
	}

	req := ctrlRuntime.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-controller",
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

	// Verify ControllerMeta was created
	controllerMeta := &k8sv1alpha1.ControllerMeta{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "controller-uid-createsmeta",
		Namespace: "default",
	}, controllerMeta)
	if err != nil {
		t.Fatalf("expected ControllerMeta to be created, got error: %v", err)
	}

	if controllerMeta.Spec.ControllerName != "test-controller" {
		t.Errorf("expected ControllerMeta.Spec.ControllerName to be 'test-controller', got %q", controllerMeta.Spec.ControllerName)
	}
}

func TestControllerReconciler_MetaFoundGenerationChanged(t *testing.T) {
	scheme := newTestScheme()

	ctrl := &k8sv1alpha1.Controller{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-controller",
			Namespace:  "default",
			UID:        "controller-uid-genchange",
			Generation: 2, // Generation changed
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.ControllerSpec{
			Tenant: "admin",
			Site:   "dc1",
		},
	}

	controllerMeta := &k8sv1alpha1.ControllerMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "controller-uid-genchange",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.ControllerMetaSpec{
			ControllerCRGeneration: 1, // Old generation
			Imported:               false,
			Reclaim:                false,
			ControllerName:         "test-controller",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, ctrl, controllerMeta)

	// Create test storage with sites and tenants
	testStorage := newTestStorage([]*site.Site{
		{ID: 1, Name: "dc1"},
	})
	testStorage.TenantsStorage.Tenants = []*tenant.Tenant{
		{ID: 1, Name: "admin"},
	}

	r := &ControllerReconciler{
		Client:          fakeClient,
		Log:             newTestLogger(),
		Scheme:          scheme,
		InventoryClient: &MockInventoryClient{},
		NStorage:        testStorage,
	}

	req := ctrlRuntime.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-controller",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(req)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.RequeueAfter == 0 {
		t.Error("expected requeue with interval after updating meta")
	}

	// Verify ControllerMeta was updated with new generation
	updatedMeta := &k8sv1alpha1.ControllerMeta{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "controller-uid-genchange",
		Namespace: "default",
	}, updatedMeta)
	if err != nil {
		t.Fatalf("failed to get updated ControllerMeta: %v", err)
	}

	if updatedMeta.Spec.ControllerCRGeneration != 2 {
		t.Errorf("expected ControllerMeta.Spec.ControllerCRGeneration to be 2, got %d", updatedMeta.Spec.ControllerCRGeneration)
	}
}
