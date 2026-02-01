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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestLinkMetaReconciler_LinkMetaNotFound(t *testing.T) {
	scheme := newTestScheme()
	fakeClient := fake.NewFakeClientWithScheme(scheme)

	r := &LinkMetaReconciler{
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

func TestLinkMetaReconciler_DeletionWithReclaim_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	linkMeta := &k8sv1alpha1.LinkMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "link-meta-reclaim",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.LinkMetaSpec{
			ID:       "100",
			Reclaim:  true,
			LinkName: "test-link",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, linkMeta)

	r := &LinkMetaReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "link-meta-reclaim",
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

func TestLinkMetaReconciler_DeletionWithEmptyID_SkipsAPICall(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	linkMeta := &k8sv1alpha1.LinkMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "link-meta-empty",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.LinkMetaSpec{
			ID:       "",
			Reclaim:  false,
			LinkName: "test-link",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, linkMeta)

	r := &LinkMetaReconciler{
		Client: fakeClient,
		Log:    newTestLogger(),
		Scheme: scheme,
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "link-meta-empty",
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

func TestLinkMetaReconciler_CreateLink(t *testing.T) {
	scheme := newTestScheme()

	linkMeta := &k8sv1alpha1.LinkMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "link-meta-create",
			Namespace:  "default",
			Finalizers: []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.LinkMetaSpec{
			ID:       "", // No ID means create
			LinkName: "test-link",
			Local:    100,
			Remote:   200,
		},
	}

	linkCR := &k8sv1alpha1.Link{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-link",
			Namespace: "default",
		},
		Spec: k8sv1alpha1.LinkSpec{
			Ports: []k8sv1alpha1.LinkSpecPort{"switch1@port1", "switch2@port2"},
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, linkMeta, linkCR)

	r := &LinkMetaReconciler{
		Client:     fakeClient,
		Log:        newTestLogger(),
		Scheme:     scheme,
		NStorage:   newTestStorage(nil),
		LinkClient: &MockLinkClient{},
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "link-meta-create",
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

	// Verify ID was set after creation
	updated := &k8sv1alpha1.LinkMeta{}
	err = fakeClient.Get(context.Background(), req.NamespacedName, updated)
	if err != nil {
		t.Fatalf("failed to get updated LinkMeta: %v", err)
	}

	if updated.Spec.ID == "" {
		t.Errorf("expected ID to be set after creation, got empty string")
	}
}

func TestLinkMetaReconciler_DeletionWithID(t *testing.T) {
	scheme := newTestScheme()

	now := metav1.Now()
	linkMeta := &k8sv1alpha1.LinkMeta{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "link-meta-delete",
			Namespace:         "default",
			DeletionTimestamp: &now,
			Finalizers:        []string{"resource.k8s.netris.ai/delete"},
		},
		Spec: k8sv1alpha1.LinkMetaSpec{
			ID:       "100-200",
			Reclaim:  false, // Do not reclaim, actually delete
			LinkName: "test-link",
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(scheme, linkMeta)

	r := &LinkMetaReconciler{
		Client:     fakeClient,
		Log:        newTestLogger(),
		Scheme:     scheme,
		LinkClient: &MockLinkClient{},
	}

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "link-meta-delete",
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
