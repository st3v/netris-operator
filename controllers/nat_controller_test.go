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
	"github.com/netrisai/netriswebapi/v2/types/site"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNatReconciler_NatNotFound(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	r := &NatReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent-nat",
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

func TestNatReconciler_SetsDefaultAnnotations(t *testing.T) {
	scheme := newTestScheme()

	nat := &k8sv1alpha1.Nat{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-nat",
			Namespace: "default",
			UID:       "nat-uid-123",
			Annotations: map[string]string{
				"test": "placeholder",
			},
		},
		Spec: k8sv1alpha1.NatSpec{
			Site:       "dc1",
			Action:     "snat",
			Protocol:   "tcp",
			SrcAddress: "10.0.0.0/24",
			DstAddress: "0.0.0.0/0",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, nat)

	r := &NatReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-nat",
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
	updated := &k8sv1alpha1.Nat{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated Nat: %v", err)
	}

	if updated.GetAnnotations()["resource.k8s.netris.ai/import"] != "false" {
		t.Errorf("expected import annotation 'false', got %q", updated.GetAnnotations()["resource.k8s.netris.ai/import"])
	}
	if updated.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"] != "delete" {
		t.Errorf("expected reclaimPolicy 'delete', got %q", updated.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"])
	}
}

func TestNatReconciler_SetsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	nat := &k8sv1alpha1.Nat{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-nat",
			Namespace: "default",
			UID:       "nat-uid-456",
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
		},
		Spec: k8sv1alpha1.NatSpec{
			Site:       "dc1",
			Action:     "snat",
			Protocol:   "tcp",
			SrcAddress: "10.0.0.0/24",
			DstAddress: "0.0.0.0/0",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, nat)

	r := &NatReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-nat",
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
	updated := &k8sv1alpha1.Nat{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated Nat: %v", err)
	}

	if len(updated.GetFinalizers()) == 0 || updated.GetFinalizers()[0] != "resource.k8s.netris.ai/delete" {
		t.Errorf("expected finalizer to be set, got %v", updated.GetFinalizers())
	}
}

func TestNatReconciler_MetaFoundNoChanges(t *testing.T) {
	scheme := newTestScheme()

	nat := &k8sv1alpha1.Nat{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-nat",
			Namespace:  "default",
			UID:        "nat-uid-abc",
			Generation: 1,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.NatSpec{
			Site:       "dc1",
			Action:     "snat",
			Protocol:   "tcp",
			SrcAddress: "10.0.0.0/24",
			DstAddress: "0.0.0.0/0",
		},
	}

	natMeta := &k8sv1alpha1.NatMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nat-uid-abc",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.NatMetaSpec{
			NatCRGeneration: 1, // matches
			Imported:        false,
			Reclaim:         false,
			NatName:         "test-nat",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, nat, natMeta)

	r := &NatReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-nat",
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

func TestNatReconciler_DeletionNoMeta_ClearsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	nat := &k8sv1alpha1.Nat{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-nat",
			Namespace:         "default",
			UID:               "nat-uid-nometa",
			DeletionTimestamp: &now,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.NatSpec{
			Site:       "dc1",
			Action:     "snat",
			Protocol:   "tcp",
			SrcAddress: "10.0.0.0/24",
			DstAddress: "0.0.0.0/0",
		},
	}

	// No NatMeta exists
	fakeClient := fake.NewFakeClientWithScheme(scheme, nat)

	r := &NatReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-nat",
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
	updated := &k8sv1alpha1.Nat{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated Nat: %v", err)
	}

	if len(updated.GetFinalizers()) != 0 {
		t.Errorf("expected finalizers to be cleared, got %v", updated.GetFinalizers())
	}
}

func TestNatReconciler_DeletionWithReclaim_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	nat := &k8sv1alpha1.Nat{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-nat",
			Namespace:         "default",
			UID:               "nat-uid-reclaim",
			DeletionTimestamp: &now,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "true",
				"resource.k8s.netris.ai/reclaimPolicy": "retain",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.NatSpec{
			Site:       "dc1",
			Action:     "snat",
			Protocol:   "tcp",
			SrcAddress: "10.0.0.0/24",
			DstAddress: "0.0.0.0/0",
		},
	}

	natMeta := &k8sv1alpha1.NatMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "nat-uid-reclaim",
			Namespace:  "default",
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.NatMetaSpec{
			ID:      100,
			Reclaim: true, // Skip API call
			NatName: "test-nat",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, nat, natMeta)

	r := &NatReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
		// Cred is nil - if API call were attempted, it would panic
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-nat",
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

func TestNatReconciler_CreatesMeta(t *testing.T) {
	scheme := newTestScheme()

	nat := &k8sv1alpha1.Nat{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-nat",
			Namespace:  "default",
			UID:        "nat-uid-createsmeta",
			Generation: 1,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.NatSpec{
			Site:       "dc1",
			Action:     "snat",
			SrcAddress: "10.0.0.0/24",
			DstAddress: "0.0.0.0/0",
			SnatToIP:   "192.168.1.1",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, nat)

	// Create test storage with sites
	testStorage := newTestStorage([]*site.Site{
		{ID: 1, Name: "dc1"},
	})

	r := &NatReconciler{
		Client:    fakeClient,
		Log:       newTestLogger(),
		Scheme:    scheme,
		NATClient: &MockNATClient{},
		NStorage:  testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-nat",
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

	// Verify NatMeta was created
	natMeta := &k8sv1alpha1.NatMeta{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "nat-uid-createsmeta",
		Namespace: "default",
	}, natMeta)
	if err != nil {
		t.Fatalf("expected NatMeta to be created, got error: %v", err)
	}

	if natMeta.Spec.NatName != "test-nat" {
		t.Errorf("expected NatMeta.Spec.NatName to be 'test-nat', got %q", natMeta.Spec.NatName)
	}
}

func TestNatReconciler_MetaFoundGenerationChanged(t *testing.T) {
	scheme := newTestScheme()

	nat := &k8sv1alpha1.Nat{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-nat",
			Namespace:  "default",
			UID:        "nat-uid-genchange",
			Generation: 2, // Generation changed
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.NatSpec{
			Site:       "dc1",
			Action:     "snat",
			SrcAddress: "10.0.0.0/24",
			DstAddress: "0.0.0.0/0",
			SnatToIP:   "192.168.1.1",
		},
	}

	natMeta := &k8sv1alpha1.NatMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nat-uid-genchange",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.NatMetaSpec{
			NatCRGeneration: 1, // Old generation
			Imported:        false,
			Reclaim:         false,
			NatName:         "test-nat",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, nat, natMeta)

	// Create test storage with sites
	testStorage := newTestStorage(nil)
	testStorage.SitesStorage.Sites = []*site.Site{
		{ID: 1, Name: "dc1"},
	}

	r := &NatReconciler{
		Client:    fakeClient,
		Log:       newTestLogger(),
		Scheme:    scheme,
		NATClient: &MockNATClient{},
		NStorage:  testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-nat",
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

	// Verify NatMeta was updated with new generation
	updatedMeta := &k8sv1alpha1.NatMeta{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "nat-uid-genchange",
		Namespace: "default",
	}, updatedMeta)
	if err != nil {
		t.Fatalf("failed to get updated NatMeta: %v", err)
	}

	if updatedMeta.Spec.NatCRGeneration != 2 {
		t.Errorf("expected NatMeta.Spec.NatCRGeneration to be 2, got %d", updatedMeta.Spec.NatCRGeneration)
	}
}
