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
	"github.com/netrisai/netriswebapi/v2/types/inventory"
	"github.com/netrisai/netriswebapi/v2/types/site"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSwitchReconciler_SwitchNotFound(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	r := &SwitchReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent-switch",
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

func TestSwitchReconciler_SetsDefaultAnnotations(t *testing.T) {
	scheme := newTestScheme()

	sw := &k8sv1alpha1.Switch{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-switch",
			Namespace: "default",
			UID:       "switch-uid-123",
			Annotations: map[string]string{
				"test": "placeholder",
			},
		},
		Spec: k8sv1alpha1.SwitchSpec{
			Tenant: "admin",
			Site:   "dc1",
			NOS:    "cumulus_linux",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, sw)

	r := &SwitchReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-switch",
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

	updated := &k8sv1alpha1.Switch{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated Switch: %v", err)
	}

	if updated.GetAnnotations()["resource.k8s.netris.ai/import"] != "false" {
		t.Errorf("expected import annotation 'false', got %q", updated.GetAnnotations()["resource.k8s.netris.ai/import"])
	}
	if updated.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"] != "delete" {
		t.Errorf("expected reclaimPolicy 'delete', got %q", updated.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"])
	}
}

func TestSwitchReconciler_SetsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	sw := &k8sv1alpha1.Switch{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-switch",
			Namespace: "default",
			UID:       "switch-uid-456",
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
		},
		Spec: k8sv1alpha1.SwitchSpec{
			Tenant: "admin",
			Site:   "dc1",
			NOS:    "cumulus_linux",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, sw)

	r := &SwitchReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-switch",
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

	updated := &k8sv1alpha1.Switch{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated Switch: %v", err)
	}

	if len(updated.GetFinalizers()) == 0 || updated.GetFinalizers()[0] != "resource.k8s.netris.ai/delete" {
		t.Errorf("expected finalizer to be set, got %v", updated.GetFinalizers())
	}
}

func TestSwitchReconciler_MetaFoundNoChanges(t *testing.T) {
	scheme := newTestScheme()

	sw := &k8sv1alpha1.Switch{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-switch",
			Namespace:  "default",
			UID:        "switch-uid-abc",
			Generation: 1,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SwitchSpec{
			Tenant: "admin",
			Site:   "dc1",
			NOS:    "cumulus_linux",
		},
	}

	switchMeta := &k8sv1alpha1.SwitchMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "switch-uid-abc",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.SwitchMetaSpec{
			SwitchCRGeneration: 1,
			Imported:           false,
			Reclaim:            false,
			SwitchName:         "test-switch",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, sw, switchMeta)

	r := &SwitchReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-switch",
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

func TestSwitchReconciler_DeletionNoMeta_ClearsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	sw := &k8sv1alpha1.Switch{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-switch",
			Namespace:         "default",
			UID:               "switch-uid-nometa",
			DeletionTimestamp: &now,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SwitchSpec{
			Tenant: "admin",
			Site:   "dc1",
			NOS:    "cumulus_linux",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, sw)

	r := &SwitchReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-switch",
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

	updated := &k8sv1alpha1.Switch{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated Switch: %v", err)
	}

	if len(updated.GetFinalizers()) != 0 {
		t.Errorf("expected finalizers to be cleared, got %v", updated.GetFinalizers())
	}
}

func TestSwitchReconciler_DeletionWithReclaim_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	sw := &k8sv1alpha1.Switch{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-switch",
			Namespace:         "default",
			UID:               "switch-uid-reclaim",
			DeletionTimestamp: &now,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "true",
				"resource.k8s.netris.ai/reclaimPolicy": "retain",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SwitchSpec{
			Tenant: "admin",
			Site:   "dc1",
			NOS:    "cumulus_linux",
		},
	}

	switchMeta := &k8sv1alpha1.SwitchMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "switch-uid-reclaim",
			Namespace:  "default",
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SwitchMetaSpec{
			ID:         100,
			Reclaim:    true,
			SwitchName: "test-switch",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, sw, switchMeta)

	r := &SwitchReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-switch",
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

func TestSwitchReconciler_CreatesMeta(t *testing.T) {
	scheme := newTestScheme()

	sw := &k8sv1alpha1.Switch{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-switch",
			Namespace:  "default",
			UID:        "switch-uid-createsmeta",
			Generation: 1,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SwitchSpec{
			Tenant:     "admin",
			Site:       "dc1",
			NOS:        "cumulus_linux",
			PortsCount: 48,
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, sw)

	// Create test storage with sites and tenants
	testStorage := newTestStorage([]*site.Site{
		{ID: 1, Name: "dc1"},
	})
	testStorage.TenantsStorage.Tenants = []*tenant.Tenant{
		{ID: 1, Name: "admin"},
	}

	r := &SwitchReconciler{
		Client:                 fakeClient,
		Log:                    newTestLogger(),
		Scheme:                 scheme,
		InventoryClient:        &MockInventoryClient{NOSData: []*inventory.NOS{{Tag: "cumulus_linux"}}},
		InventoryProfileClient: &MockInventoryProfileClient{},
		NStorage:               testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-switch",
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

	// Verify SwitchMeta was created
	switchMeta := &k8sv1alpha1.SwitchMeta{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "switch-uid-createsmeta",
		Namespace: "default",
	}, switchMeta)
	if err != nil {
		t.Fatalf("expected SwitchMeta to be created, got error: %v", err)
	}

	if switchMeta.Spec.SwitchName != "test-switch" {
		t.Errorf("expected SwitchMeta.Spec.SwitchName to be 'test-switch', got %q", switchMeta.Spec.SwitchName)
	}
}

func TestSwitchReconciler_MetaFoundGenerationChanged(t *testing.T) {
	scheme := newTestScheme()

	sw := &k8sv1alpha1.Switch{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-switch",
			Namespace:  "default",
			UID:        "switch-uid-genchange",
			Generation: 2, // Generation changed
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SwitchSpec{
			Tenant:     "admin",
			Site:       "dc1",
			NOS:        "cumulus_linux",
			PortsCount: 48,
		},
	}

	switchMeta := &k8sv1alpha1.SwitchMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "switch-uid-genchange",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.SwitchMetaSpec{
			SwitchCRGeneration: 1, // Old generation
			Imported:           false,
			Reclaim:            false,
			SwitchName:         "test-switch",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, sw, switchMeta)

	// Create test storage with sites and tenants
	testStorage := newTestStorage([]*site.Site{
		{ID: 1, Name: "dc1"},
	})
	testStorage.TenantsStorage.Tenants = []*tenant.Tenant{
		{ID: 1, Name: "admin"},
	}

	r := &SwitchReconciler{
		Client:                 fakeClient,
		Log:                    newTestLogger(),
		Scheme:                 scheme,
		InventoryClient:        &MockInventoryClient{NOSData: []*inventory.NOS{{Tag: "cumulus_linux"}}},
		InventoryProfileClient: &MockInventoryProfileClient{},
		NStorage:               testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-switch",
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

	// Verify SwitchMeta was updated with new generation
	updatedMeta := &k8sv1alpha1.SwitchMeta{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "switch-uid-genchange",
		Namespace: "default",
	}, updatedMeta)
	if err != nil {
		t.Fatalf("failed to get updated SwitchMeta: %v", err)
	}

	if updatedMeta.Spec.SwitchCRGeneration != 2 {
		t.Errorf("expected SwitchMeta.Spec.SwitchCRGeneration to be 2, got %d", updatedMeta.Spec.SwitchCRGeneration)
	}
}
