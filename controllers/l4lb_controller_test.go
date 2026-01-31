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

func TestL4LBReconciler_L4LBNotFound(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	r := &L4LBReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent-l4lb",
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

func TestL4LBReconciler_SetsDefaultAnnotations(t *testing.T) {
	scheme := newTestScheme()

	l4lb := &k8sv1alpha1.L4LB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-l4lb",
			Namespace: "default",
			UID:       "l4lb-uid-123",
			Annotations: map[string]string{
				"test": "placeholder",
			},
		},
		Spec: k8sv1alpha1.L4LBSpec{
			Site:     "dc1",
			Protocol: "tcp",
			Frontend: k8sv1alpha1.L4LBFrontend{
				Port: 80,
				IP:   "10.0.0.100",
			},
			Backend: []k8sv1alpha1.L4LBBackend{
				"10.0.0.1:8080",
			},
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, l4lb)

	r := &L4LBReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-l4lb",
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

	// Verify annotations were set
	updated := &k8sv1alpha1.L4LB{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated L4LB: %v", err)
	}

	if updated.GetAnnotations()["resource.k8s.netris.ai/import"] != "false" {
		t.Errorf("expected import annotation 'false', got %q", updated.GetAnnotations()["resource.k8s.netris.ai/import"])
	}
	if updated.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"] != "delete" {
		t.Errorf("expected reclaimPolicy 'delete', got %q", updated.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"])
	}
}

func TestL4LBReconciler_SetsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	l4lb := &k8sv1alpha1.L4LB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-l4lb",
			Namespace: "default",
			UID:       "l4lb-uid-456",
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
		},
		Spec: k8sv1alpha1.L4LBSpec{
			Site:     "dc1",
			Protocol: "tcp",
			Frontend: k8sv1alpha1.L4LBFrontend{
				Port: 80,
				IP:   "10.0.0.100",
			},
			Backend: []k8sv1alpha1.L4LBBackend{
				"10.0.0.1:8080",
			},
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, l4lb)

	r := &L4LBReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-l4lb",
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
	updated := &k8sv1alpha1.L4LB{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated L4LB: %v", err)
	}

	if len(updated.GetFinalizers()) == 0 || updated.GetFinalizers()[0] != "resource.k8s.netris.ai/delete" {
		t.Errorf("expected finalizer to be set, got %v", updated.GetFinalizers())
	}
}

func TestL4LBReconciler_MetaFoundNoChanges(t *testing.T) {
	scheme := newTestScheme()

	l4lb := &k8sv1alpha1.L4LB{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-l4lb",
			Namespace:  "default",
			UID:        "l4lb-uid-abc",
			Generation: 1,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.L4LBSpec{
			Site:     "dc1",
			Protocol: "tcp",
			Frontend: k8sv1alpha1.L4LBFrontend{
				Port: 80,
				IP:   "10.0.0.100",
			},
			Backend: []k8sv1alpha1.L4LBBackend{
				"10.0.0.1:8080",
			},
		},
	}

	l4lbMeta := &k8sv1alpha1.L4LBMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "l4lb-uid-abc",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.L4LBMetaSpec{
			L4LBCRGeneration: 1, // matches
			Imported:         false,
			Reclaim:          false,
			L4LBName:         "test-l4lb",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, l4lb, l4lbMeta)

	r := &L4LBReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-l4lb",
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

func TestL4LBReconciler_DeletionNoMeta_ClearsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	l4lb := &k8sv1alpha1.L4LB{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-l4lb",
			Namespace:         "default",
			UID:               "l4lb-uid-nometa",
			DeletionTimestamp: &now,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.L4LBSpec{
			Site:     "dc1",
			Protocol: "tcp",
			Frontend: k8sv1alpha1.L4LBFrontend{
				Port: 80,
				IP:   "10.0.0.100",
			},
			Backend: []k8sv1alpha1.L4LBBackend{
				"10.0.0.1:8080",
			},
		},
	}

	// No L4LBMeta exists
	fakeClient := fake.NewFakeClientWithScheme(scheme, l4lb)

	r := &L4LBReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-l4lb",
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

	// Verify finalizer was cleared
	updated := &k8sv1alpha1.L4LB{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated L4LB: %v", err)
	}

	if len(updated.GetFinalizers()) != 0 {
		t.Errorf("expected finalizers to be cleared, got %v", updated.GetFinalizers())
	}
}

func TestL4LBReconciler_DeletionWithReclaim_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	l4lb := &k8sv1alpha1.L4LB{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-l4lb",
			Namespace:         "default",
			UID:               "l4lb-uid-reclaim",
			DeletionTimestamp: &now,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "true",
				"resource.k8s.netris.ai/reclaimPolicy": "retain",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.L4LBSpec{
			Site:     "dc1",
			Protocol: "tcp",
			Frontend: k8sv1alpha1.L4LBFrontend{
				Port: 80,
				IP:   "10.0.0.100",
			},
			Backend: []k8sv1alpha1.L4LBBackend{
				"10.0.0.1:8080",
			},
		},
	}

	l4lbMeta := &k8sv1alpha1.L4LBMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "l4lb-uid-reclaim",
			Namespace:  "default",
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.L4LBMetaSpec{
			ID:       100,
			Reclaim:  true, // Skip API call
			L4LBName: "test-l4lb",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, l4lb, l4lbMeta)

	r := &L4LBReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
		// Cred is nil - if API call were attempted, it would panic
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-l4lb",
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

func TestL4LBReconciler_CreatesMeta(t *testing.T) {
	scheme := newTestScheme()

	l4lb := &k8sv1alpha1.L4LB{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-l4lb",
			Namespace:  "default",
			UID:        "l4lb-uid-createsmeta",
			Generation: 1,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.L4LBSpec{
			OwnerTenant: "admin",
			Site:        "dc1",
			Frontend: k8sv1alpha1.L4LBFrontend{
				IP:   "10.0.0.1",
				Port: 80,
			},
			Backend:  []k8sv1alpha1.L4LBBackend{"10.0.0.10:8080"},
			Protocol: "tcp",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, l4lb)

	// Create test storage with sites and tenants
	testStorage := newTestStorage([]*site.Site{
		{ID: 1, Name: "dc1"},
	})
	testStorage.TenantsStorage.Tenants = []*tenant.Tenant{
		{ID: 1, Name: "admin"},
	}

	r := &L4LBReconciler{
		Client:     fakeClient,
		Log:        newTestLogger(),
		Scheme:     scheme,
		L4LBClient: &MockL4LBClient{},
		IPAMClient: &MockIPAMClient{},
		NStorage:   testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-l4lb",
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

	// Verify L4LBMeta was created
	l4lbMeta := &k8sv1alpha1.L4LBMeta{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "l4lb-uid-createsmeta",
		Namespace: "default",
	}, l4lbMeta)
	if err != nil {
		t.Fatalf("expected L4LBMeta to be created, got error: %v", err)
	}

	if l4lbMeta.Spec.L4LBName != "test-l4lb" {
		t.Errorf("expected L4LBMeta.Spec.L4LBName to be 'test-l4lb', got %q", l4lbMeta.Spec.L4LBName)
	}
}

func TestL4LBReconciler_MetaFoundGenerationChanged(t *testing.T) {
	scheme := newTestScheme()

	l4lb := &k8sv1alpha1.L4LB{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-l4lb",
			Namespace:  "default",
			UID:        "l4lb-uid-genchange",
			Generation: 2, // Generation changed
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.L4LBSpec{
			OwnerTenant: "admin",
			Site:        "dc1",
			Frontend: k8sv1alpha1.L4LBFrontend{
				IP:   "10.0.0.1",
				Port: 80,
			},
			Backend:  []k8sv1alpha1.L4LBBackend{"10.0.0.10:8080"},
			Protocol: "tcp",
		},
	}

	l4lbMeta := &k8sv1alpha1.L4LBMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "l4lb-uid-genchange",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.L4LBMetaSpec{
			L4LBCRGeneration: 1, // Old generation
			Imported:         false,
			Reclaim:          false,
			L4LBName:         "test-l4lb",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, l4lb, l4lbMeta)

	// Create test storage with sites and tenants
	testStorage := newTestStorage([]*site.Site{
		{ID: 1, Name: "dc1"},
	})
	testStorage.TenantsStorage.Tenants = []*tenant.Tenant{
		{ID: 1, Name: "admin"},
	}

	r := &L4LBReconciler{
		Client:     fakeClient,
		Log:        newTestLogger(),
		Scheme:     scheme,
		L4LBClient: &MockL4LBClient{},
		IPAMClient: &MockIPAMClient{},
		NStorage:   testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-l4lb",
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

	// Verify L4LBMeta was updated with new generation
	updatedMeta := &k8sv1alpha1.L4LBMeta{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "l4lb-uid-genchange",
		Namespace: "default",
	}, updatedMeta)
	if err != nil {
		t.Fatalf("failed to get updated L4LBMeta: %v", err)
	}

	if updatedMeta.Spec.L4LBCRGeneration != 2 {
		t.Errorf("expected L4LBMeta.Spec.L4LBCRGeneration to be 2, got %d", updatedMeta.Spec.L4LBCRGeneration)
	}
}
