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
	"github.com/netrisai/netriswebapi/v2/types/site"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestBGPReconciler_BGPNotFound(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	r := &BGPReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent-bgp",
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

func TestBGPReconciler_SetsDefaultAnnotations(t *testing.T) {
	scheme := newTestScheme()

	bgp := &k8sv1alpha1.BGP{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-bgp",
			Namespace: "default",
			UID:       "bgp-uid-123",
			Annotations: map[string]string{
				"test": "placeholder",
			},
		},
		Spec: k8sv1alpha1.BGPSpec{
			Site:       "dc1",
			NeighborAS: 65001,
			LocalIP:    "10.0.0.1",
			RemoteIP:   "10.0.0.2",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, bgp)

	r := &BGPReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-bgp",
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
	updated := &k8sv1alpha1.BGP{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated BGP: %v", err)
	}

	if updated.GetAnnotations()["resource.k8s.netris.ai/import"] != "false" {
		t.Errorf("expected import annotation 'false', got %q", updated.GetAnnotations()["resource.k8s.netris.ai/import"])
	}
	if updated.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"] != "delete" {
		t.Errorf("expected reclaimPolicy 'delete', got %q", updated.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"])
	}
}

func TestBGPReconciler_SetsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	bgp := &k8sv1alpha1.BGP{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-bgp",
			Namespace: "default",
			UID:       "bgp-uid-456",
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
		},
		Spec: k8sv1alpha1.BGPSpec{
			Site:       "dc1",
			NeighborAS: 65001,
			LocalIP:    "10.0.0.1",
			RemoteIP:   "10.0.0.2",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, bgp)

	r := &BGPReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-bgp",
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
	updated := &k8sv1alpha1.BGP{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated BGP: %v", err)
	}

	if len(updated.GetFinalizers()) == 0 || updated.GetFinalizers()[0] != "resource.k8s.netris.ai/delete" {
		t.Errorf("expected finalizer to be set, got %v", updated.GetFinalizers())
	}
}

// TestBGPReconciler_CreatesMeta tests that when a BGP has a finalizer but no BGPMeta,
// the reconciler creates a BGPMeta resource.
func TestBGPReconciler_CreatesMeta(t *testing.T) {
	scheme := newTestScheme()

	bgp := &k8sv1alpha1.BGP{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-bgp",
			Namespace:  "default",
			UID:        "bgp-uid-createsmeta",
			Generation: 1,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.BGPSpec{
			Site:       "dc1",
			NeighborAS: 65001,
			LocalIP:    "10.0.0.1/24",
			RemoteIP:   "10.0.0.2/24",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, bgp)

	// Create test storage with sites
	testStorage := newTestStorage([]*site.Site{
		{ID: 1, Name: "dc1"},
	})

	r := &BGPReconciler{
		Client:          fakeClient,
		Log:             newTestLogger(),
		Scheme:          scheme,
		BGPClient:       &MockBGPClient{},
		InventoryClient: &MockInventoryClient{HWData: []*inventory.HW{{ID: 1, Name: "switch1"}}},
		VNetClient:      &MockVNetClient{},
		NStorage:        testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-bgp",
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

	// Verify BGPMeta was created
	bgpMeta := &k8sv1alpha1.BGPMeta{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "bgp-uid-createsmeta",
		Namespace: "default",
	}, bgpMeta)
	if err != nil {
		t.Fatalf("expected BGPMeta to be created, got error: %v", err)
	}

	if bgpMeta.Spec.BGPName != "test-bgp" {
		t.Errorf("expected BGPMeta.Spec.BGPName to be 'test-bgp', got %q", bgpMeta.Spec.BGPName)
	}
	if bgpMeta.Spec.Site != "dc1" {
		t.Errorf("expected BGPMeta.Spec.Site to be 'dc1', got %q", bgpMeta.Spec.Site)
	}
	if bgpMeta.Spec.BGPCRGeneration != 1 {
		t.Errorf("expected BGPMeta.Spec.BGPCRGeneration to be 1, got %d", bgpMeta.Spec.BGPCRGeneration)
	}
}

func TestBGPReconciler_MetaFoundNoChanges(t *testing.T) {
	scheme := newTestScheme()

	bgp := &k8sv1alpha1.BGP{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-bgp",
			Namespace:  "default",
			UID:        "bgp-uid-abc",
			Generation: 1,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.BGPSpec{
			Site:       "dc1",
			NeighborAS: 65001,
			LocalIP:    "10.0.0.1",
			RemoteIP:   "10.0.0.2",
		},
	}

	bgpMeta := &k8sv1alpha1.BGPMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bgp-uid-abc",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.BGPMetaSpec{
			BGPCRGeneration: 1, // matches
			Imported:        false,
			Reclaim:         false,
			BGPName:         "test-bgp",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, bgp, bgpMeta)

	r := &BGPReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-bgp",
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

// TestBGPReconciler_MetaFoundGenerationChanged tests that when the BGP generation changes,
// the reconciler updates the BGPMeta with new values.
func TestBGPReconciler_MetaFoundGenerationChanged(t *testing.T) {
	scheme := newTestScheme()

	bgp := &k8sv1alpha1.BGP{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-bgp",
			Namespace:  "default",
			UID:        "bgp-uid-genchange",
			Generation: 2, // Generation changed from 1 to 2
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.BGPSpec{
			Site:        "dc1",
			NeighborAS:  65002, // Changed from 65001
			LocalIP:     "10.0.0.1/24",
			RemoteIP:    "10.0.0.2/24",
			Description: "updated description",
		},
	}

	bgpMeta := &k8sv1alpha1.BGPMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bgp-uid-genchange",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.BGPMetaSpec{
			ID:              100,
			BGPCRGeneration: 1, // Old generation
			Imported:        false,
			Reclaim:         false,
			BGPName:         "test-bgp",
			Site:            "dc1",
			NeighborAs:      65001, // Old value
			Description:     "old description",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, bgp, bgpMeta)

	// Create test storage with sites
	testStorage := newTestStorage([]*site.Site{
		{ID: 1, Name: "dc1"},
	})

	r := &BGPReconciler{
		Client:          fakeClient,
		Log:             newTestLogger(),
		Scheme:          scheme,
		BGPClient:       &MockBGPClient{},
		InventoryClient: &MockInventoryClient{HWData: []*inventory.HW{{ID: 1, Name: "switch1"}}},
		VNetClient:      &MockVNetClient{},
		NStorage:        testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-bgp",
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

	// Verify BGPMeta was updated
	updatedMeta := &k8sv1alpha1.BGPMeta{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "bgp-uid-genchange",
		Namespace: "default",
	}, updatedMeta)
	if err != nil {
		t.Fatalf("failed to get updated BGPMeta: %v", err)
	}

	// ID should be preserved
	if updatedMeta.Spec.ID != 100 {
		t.Errorf("expected BGPMeta.Spec.ID to be preserved as 100, got %d", updatedMeta.Spec.ID)
	}
	// Generation should be updated
	if updatedMeta.Spec.BGPCRGeneration != 2 {
		t.Errorf("expected BGPMeta.Spec.BGPCRGeneration to be 2, got %d", updatedMeta.Spec.BGPCRGeneration)
	}
	// NeighborAs should be updated
	if updatedMeta.Spec.NeighborAs != 65002 {
		t.Errorf("expected BGPMeta.Spec.NeighborAs to be 65002, got %d", updatedMeta.Spec.NeighborAs)
	}
	// Description should be updated
	if updatedMeta.Spec.Description != "updated description" {
		t.Errorf("expected BGPMeta.Spec.Description to be 'updated description', got %q", updatedMeta.Spec.Description)
	}
}

func TestBGPReconciler_DeletionNoMeta_ClearsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	bgp := &k8sv1alpha1.BGP{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-bgp",
			Namespace:         "default",
			UID:               "bgp-uid-nometa",
			DeletionTimestamp: &now,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.BGPSpec{
			Site:       "dc1",
			NeighborAS: 65001,
			LocalIP:    "10.0.0.1",
			RemoteIP:   "10.0.0.2",
		},
	}

	// No BGPMeta exists
	fakeClient := fake.NewFakeClientWithScheme(scheme, bgp)

	r := &BGPReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-bgp",
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
	updated := &k8sv1alpha1.BGP{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated BGP: %v", err)
	}

	if len(updated.GetFinalizers()) != 0 {
		t.Errorf("expected finalizers to be cleared, got %v", updated.GetFinalizers())
	}
}

func TestBGPReconciler_DeletionWithReclaim_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	bgp := &k8sv1alpha1.BGP{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-bgp",
			Namespace:         "default",
			UID:               "bgp-uid-reclaim",
			DeletionTimestamp: &now,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "true",
				"resource.k8s.netris.ai/reclaimPolicy": "retain",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.BGPSpec{
			Site:       "dc1",
			NeighborAS: 65001,
			LocalIP:    "10.0.0.1",
			RemoteIP:   "10.0.0.2",
		},
	}

	bgpMeta := &k8sv1alpha1.BGPMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "bgp-uid-reclaim",
			Namespace:  "default",
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.BGPMetaSpec{
			ID:      100,
			Reclaim: true, // Skip API call
			BGPName: "test-bgp",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, bgp, bgpMeta)

	r := &BGPReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
		// Cred is nil - if API call were attempted, it would panic
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-bgp",
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

// TestBGPReconciler_ImportedBGP tests that when a BGP is marked as imported,
// the BGPMeta correctly reflects the imported flag.
func TestBGPReconciler_ImportedBGP(t *testing.T) {
	scheme := newTestScheme()

	bgp := &k8sv1alpha1.BGP{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-bgp-imported",
			Namespace:  "default",
			UID:        "bgp-uid-imported",
			Generation: 1,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "true",
				"resource.k8s.netris.ai/reclaimPolicy": "retain",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.BGPSpec{
			Site:       "dc1",
			NeighborAS: 65001,
			LocalIP:    "10.0.0.1/24",
			RemoteIP:   "10.0.0.2/24",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, bgp)

	// Create test storage with sites
	testStorage := newTestStorage([]*site.Site{
		{ID: 1, Name: "dc1"},
	})

	r := &BGPReconciler{
		Client:          fakeClient,
		Log:             newTestLogger(),
		Scheme:          scheme,
		BGPClient:       &MockBGPClient{},
		InventoryClient: &MockInventoryClient{HWData: []*inventory.HW{{ID: 1, Name: "switch1"}}},
		VNetClient:      &MockVNetClient{},
		NStorage:        testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-bgp-imported",
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

	// Verify BGPMeta was created with correct imported/reclaim flags
	bgpMeta := &k8sv1alpha1.BGPMeta{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "bgp-uid-imported",
		Namespace: "default",
	}, bgpMeta)
	if err != nil {
		t.Fatalf("expected BGPMeta to be created, got error: %v", err)
	}

	if !bgpMeta.Spec.Imported {
		t.Error("expected BGPMeta.Spec.Imported to be true")
	}
	if !bgpMeta.Spec.Reclaim {
		t.Error("expected BGPMeta.Spec.Reclaim to be true")
	}
	if bgpMeta.Spec.BGPName != "test-bgp-imported" {
		t.Errorf("expected BGPMeta.Spec.BGPName to be 'test-bgp-imported', got %q", bgpMeta.Spec.BGPName)
	}
}
