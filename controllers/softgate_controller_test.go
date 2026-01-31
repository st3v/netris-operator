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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSoftgateReconciler_SoftgateNotFound(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	r := &SoftgateReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent-softgate",
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

func TestSoftgateReconciler_SetsDefaultAnnotations(t *testing.T) {
	scheme := newTestScheme()

	softgate := &k8sv1alpha1.Softgate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-softgate",
			Namespace: "default",
			UID:       "softgate-uid-123",
			Annotations: map[string]string{
				"test": "placeholder",
			},
		},
		Spec: k8sv1alpha1.SoftgateSpec{
			Tenant: "admin",
			Site:   "dc1",
			MainIP: "10.0.0.1",
			MgmtIP: "192.168.1.1",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, softgate)

	r := &SoftgateReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-softgate",
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

	updated := &k8sv1alpha1.Softgate{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated Softgate: %v", err)
	}

	if updated.GetAnnotations()["resource.k8s.netris.ai/import"] != "false" {
		t.Errorf("expected import annotation 'false', got %q", updated.GetAnnotations()["resource.k8s.netris.ai/import"])
	}
	if updated.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"] != "delete" {
		t.Errorf("expected reclaimPolicy 'delete', got %q", updated.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"])
	}
}

func TestSoftgateReconciler_SetsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	softgate := &k8sv1alpha1.Softgate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-softgate",
			Namespace: "default",
			UID:       "softgate-uid-456",
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
		},
		Spec: k8sv1alpha1.SoftgateSpec{
			Tenant: "admin",
			Site:   "dc1",
			MainIP: "10.0.0.1",
			MgmtIP: "192.168.1.1",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, softgate)

	r := &SoftgateReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-softgate",
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

	updated := &k8sv1alpha1.Softgate{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated Softgate: %v", err)
	}

	if len(updated.GetFinalizers()) == 0 || updated.GetFinalizers()[0] != "resource.k8s.netris.ai/delete" {
		t.Errorf("expected finalizer to be set, got %v", updated.GetFinalizers())
	}
}

func TestSoftgateReconciler_MetaFoundNoChanges(t *testing.T) {
	scheme := newTestScheme()

	softgate := &k8sv1alpha1.Softgate{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-softgate",
			Namespace:  "default",
			UID:        "softgate-uid-abc",
			Generation: 1,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SoftgateSpec{
			Tenant: "admin",
			Site:   "dc1",
			MainIP: "10.0.0.1",
			MgmtIP: "192.168.1.1",
		},
	}

	softgateMeta := &k8sv1alpha1.SoftgateMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "softgate-uid-abc",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.SoftgateMetaSpec{
			SoftgateCRGeneration: 1,
			Imported:             false,
			Reclaim:              false,
			SoftgateName:         "test-softgate",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, softgate, softgateMeta)

	r := &SoftgateReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-softgate",
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

func TestSoftgateReconciler_DeletionNoMeta_ClearsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	softgate := &k8sv1alpha1.Softgate{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-softgate",
			Namespace:         "default",
			UID:               "softgate-uid-nometa",
			DeletionTimestamp: &now,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SoftgateSpec{
			Tenant: "admin",
			Site:   "dc1",
			MainIP: "10.0.0.1",
			MgmtIP: "192.168.1.1",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, softgate)

	r := &SoftgateReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-softgate",
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

	updated := &k8sv1alpha1.Softgate{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated Softgate: %v", err)
	}

	if len(updated.GetFinalizers()) != 0 {
		t.Errorf("expected finalizers to be cleared, got %v", updated.GetFinalizers())
	}
}

func TestSoftgateReconciler_DeletionWithReclaim_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	softgate := &k8sv1alpha1.Softgate{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-softgate",
			Namespace:         "default",
			UID:               "softgate-uid-reclaim",
			DeletionTimestamp: &now,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "true",
				"resource.k8s.netris.ai/reclaimPolicy": "retain",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SoftgateSpec{
			Tenant: "admin",
			Site:   "dc1",
			MainIP: "10.0.0.1",
			MgmtIP: "192.168.1.1",
		},
	}

	softgateMeta := &k8sv1alpha1.SoftgateMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "softgate-uid-reclaim",
			Namespace:  "default",
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SoftgateMetaSpec{
			ID:           100,
			Reclaim:      true,
			SoftgateName: "test-softgate",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, softgate, softgateMeta)

	r := &SoftgateReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-softgate",
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

func TestSoftgateReconciler_CreatesMeta(t *testing.T) {
	scheme := newTestScheme()

	sg := &k8sv1alpha1.Softgate{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-softgate",
			Namespace:  "default",
			UID:        "softgate-uid-createsmeta",
			Generation: 1,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SoftgateSpec{
			Tenant: "admin",
			Site:   "dc1",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, sg)

	// Create test storage with sites and tenants
	testStorage := newTestStorage([]*site.Site{
		{ID: 1, Name: "dc1"},
	})
	testStorage.TenantsStorage.Tenants = []*tenant.Tenant{
		{ID: 1, Name: "admin"},
	}

	r := &SoftgateReconciler{
		Client:                 fakeClient,
		Log:                    newTestLogger(),
		Scheme:                 scheme,
		InventoryClient:        &MockInventoryClient{},
		InventoryProfileClient: &MockInventoryProfileClient{},
		NStorage:               testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-softgate",
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

	// Verify SoftgateMeta was created
	softgateMeta := &k8sv1alpha1.SoftgateMeta{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "softgate-uid-createsmeta",
		Namespace: "default",
	}, softgateMeta)
	if err != nil {
		t.Fatalf("expected SoftgateMeta to be created, got error: %v", err)
	}

	if softgateMeta.Spec.SoftgateName != "test-softgate" {
		t.Errorf("expected SoftgateMeta.Spec.SoftgateName to be 'test-softgate', got %q", softgateMeta.Spec.SoftgateName)
	}
}

func TestSoftgateReconciler_MetaFoundGenerationChanged(t *testing.T) {
	scheme := newTestScheme()

	sg := &k8sv1alpha1.Softgate{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-softgate",
			Namespace:  "default",
			UID:        "softgate-uid-genchange",
			Generation: 2, // Generation changed
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.SoftgateSpec{
			Tenant: "admin",
			Site:   "dc1",
		},
	}

	softgateMeta := &k8sv1alpha1.SoftgateMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "softgate-uid-genchange",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.SoftgateMetaSpec{
			SoftgateCRGeneration: 1, // Old generation
			Imported:             false,
			Reclaim:              false,
			SoftgateName:         "test-softgate",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, sg, softgateMeta)

	// Create test storage with sites and tenants
	testStorage := newTestStorage([]*site.Site{
		{ID: 1, Name: "dc1"},
	})
	testStorage.TenantsStorage.Tenants = []*tenant.Tenant{
		{ID: 1, Name: "admin"},
	}

	r := &SoftgateReconciler{
		Client:                 fakeClient,
		Log:                    newTestLogger(),
		Scheme:                 scheme,
		InventoryClient:        &MockInventoryClient{},
		InventoryProfileClient: &MockInventoryProfileClient{},
		NStorage:               testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-softgate",
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

	// Verify SoftgateMeta was updated with new generation
	updatedMeta := &k8sv1alpha1.SoftgateMeta{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "softgate-uid-genchange",
		Namespace: "default",
	}, updatedMeta)
	if err != nil {
		t.Fatalf("failed to get updated SoftgateMeta: %v", err)
	}

	if updatedMeta.Spec.SoftgateCRGeneration != 2 {
		t.Errorf("expected SoftgateMeta.Spec.SoftgateCRGeneration to be 2, got %d", updatedMeta.Spec.SoftgateCRGeneration)
	}
}
