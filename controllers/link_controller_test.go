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
	"github.com/netrisai/netriswebapi/v2/types/port"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestLinkReconciler_LinkNotFound(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	r := &LinkReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent-link",
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

func TestLinkReconciler_SetsDefaultAnnotations(t *testing.T) {
	scheme := newTestScheme()

	link := &k8sv1alpha1.Link{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-link",
			Namespace: "default",
			UID:       "link-uid-123",
			Annotations: map[string]string{
				"test": "placeholder",
			},
		},
		Spec: k8sv1alpha1.LinkSpec{
			Ports: []k8sv1alpha1.LinkSpecPort{
				"switch1@swp1",
				"switch2@swp1",
			},
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, link)

	r := &LinkReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-link",
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

	updated := &k8sv1alpha1.Link{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated Link: %v", err)
	}

	if updated.GetAnnotations()["resource.k8s.netris.ai/import"] != "false" {
		t.Errorf("expected import annotation 'false', got %q", updated.GetAnnotations()["resource.k8s.netris.ai/import"])
	}
	if updated.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"] != "delete" {
		t.Errorf("expected reclaimPolicy 'delete', got %q", updated.GetAnnotations()["resource.k8s.netris.ai/reclaimPolicy"])
	}
}

func TestLinkReconciler_SetsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	link := &k8sv1alpha1.Link{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-link",
			Namespace: "default",
			UID:       "link-uid-456",
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
		},
		Spec: k8sv1alpha1.LinkSpec{
			Ports: []k8sv1alpha1.LinkSpecPort{
				"switch1@swp1",
				"switch2@swp1",
			},
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, link)

	r := &LinkReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-link",
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

	updated := &k8sv1alpha1.Link{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated Link: %v", err)
	}

	if len(updated.GetFinalizers()) == 0 || updated.GetFinalizers()[0] != "resource.k8s.netris.ai/delete" {
		t.Errorf("expected finalizer to be set, got %v", updated.GetFinalizers())
	}
}

func TestLinkReconciler_MetaFoundNoChanges(t *testing.T) {
	scheme := newTestScheme()

	link := &k8sv1alpha1.Link{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-link",
			Namespace:  "default",
			UID:        "link-uid-abc",
			Generation: 1,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.LinkSpec{
			Ports: []k8sv1alpha1.LinkSpecPort{
				"switch1@swp1",
				"switch2@swp1",
			},
		},
	}

	linkMeta := &k8sv1alpha1.LinkMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "link-uid-abc",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.LinkMetaSpec{
			LinkCRGeneration: 1,
			Imported:         false,
			Reclaim:          false,
			LinkName:         "test-link",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, link, linkMeta)

	r := &LinkReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-link",
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

func TestLinkReconciler_DeletionNoMeta_ClearsFinalizer(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	link := &k8sv1alpha1.Link{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-link",
			Namespace:         "default",
			UID:               "link-uid-nometa",
			DeletionTimestamp: &now,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.LinkSpec{
			Ports: []k8sv1alpha1.LinkSpecPort{
				"switch1@swp1",
				"switch2@swp1",
			},
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, link)

	r := &LinkReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-link",
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

	updated := &k8sv1alpha1.Link{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated Link: %v", err)
	}

	if len(updated.GetFinalizers()) != 0 {
		t.Errorf("expected finalizers to be cleared, got %v", updated.GetFinalizers())
	}
}

func TestLinkReconciler_DeletionWithReclaim_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	link := &k8sv1alpha1.Link{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-link",
			Namespace:         "default",
			UID:               "link-uid-reclaim",
			DeletionTimestamp: &now,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "true",
				"resource.k8s.netris.ai/reclaimPolicy": "retain",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.LinkSpec{
			Ports: []k8sv1alpha1.LinkSpecPort{
				"switch1@swp1",
				"switch2@swp1",
			},
		},
	}

	linkMeta := &k8sv1alpha1.LinkMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "link-uid-reclaim",
			Namespace:  "default",
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.LinkMetaSpec{
			ID:       "100",
			Reclaim:  true,
			LinkName: "test-link",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, link, linkMeta)

	r := &LinkReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-link",
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

func TestLinkReconciler_CreatesMeta(t *testing.T) {
	scheme := newTestScheme()

	link := &k8sv1alpha1.Link{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-link",
			Namespace:  "default",
			UID:        "link-uid-createsmeta",
			Generation: 1,
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.LinkSpec{
			Ports: []k8sv1alpha1.LinkSpecPort{"swp1@switch1", "swp2@switch2"},
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, link)

	// Create test storage with ports
	testStorage := newTestStorage(nil)
	testStorage.PortsStorage.Ports = []*port.Port{
		{ID: 1, Port_: "swp1", SwitchName: "switch1"},
		{ID: 2, Port_: "swp2", SwitchName: "switch2"},
	}

	r := &LinkReconciler{
		Client:     fakeClient,
		Log:        newTestLogger(),
		Scheme:     scheme,
		LinkClient: &MockLinkClient{},
		NStorage:   testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-link",
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

	// Verify LinkMeta was created
	linkMeta := &k8sv1alpha1.LinkMeta{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "link-uid-createsmeta",
		Namespace: "default",
	}, linkMeta)
	if err != nil {
		t.Fatalf("expected LinkMeta to be created, got error: %v", err)
	}

	if linkMeta.Spec.LinkName != "test-link" {
		t.Errorf("expected LinkMeta.Spec.LinkName to be 'test-link', got %q", linkMeta.Spec.LinkName)
	}
}

func TestLinkReconciler_MetaFoundGenerationChanged(t *testing.T) {
	scheme := newTestScheme()

	link := &k8sv1alpha1.Link{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-link",
			Namespace:  "default",
			UID:        "link-uid-genchange",
			Generation: 2, // Generation changed
			Annotations: map[string]string{
				"resource.k8s.netris.ai/import":        "false",
				"resource.k8s.netris.ai/reclaimPolicy": "delete",
			},
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.LinkSpec{
			Ports: []k8sv1alpha1.LinkSpecPort{"swp1@switch1", "swp2@switch2"},
		},
	}

	linkMeta := &k8sv1alpha1.LinkMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "link-uid-genchange",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.LinkMetaSpec{
			LinkCRGeneration: 1, // Old generation
			Imported:         false,
			Reclaim:          false,
			LinkName:         "test-link",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, link, linkMeta)

	// Create test storage with ports
	testStorage := newTestStorage(nil)
	testStorage.PortsStorage.Ports = []*port.Port{
		{ID: 1, Port_: "swp1", SwitchName: "switch1"},
		{ID: 2, Port_: "swp2", SwitchName: "switch2"},
	}

	r := &LinkReconciler{
		Client:     fakeClient,
		Log:        newTestLogger(),
		Scheme:     scheme,
		LinkClient: &MockLinkClient{},
		NStorage:   testStorage,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-link",
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

	// Verify LinkMeta was updated with new generation
	updatedMeta := &k8sv1alpha1.LinkMeta{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "link-uid-genchange",
		Namespace: "default",
	}, updatedMeta)
	if err != nil {
		t.Fatalf("failed to get updated LinkMeta: %v", err)
	}

	if updatedMeta.Spec.LinkCRGeneration != 2 {
		t.Errorf("expected LinkMeta.Spec.LinkCRGeneration to be 2, got %d", updatedMeta.Spec.LinkCRGeneration)
	}
}

